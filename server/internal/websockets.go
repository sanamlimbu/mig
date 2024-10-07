package mig

import (
	"context"
	"encoding/json"
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

// Reference - https://github.com/gorilla/websocket/blob/main/examples/chat/Client.go

const (
	writeWait      = 10 * time.Second    // time allowed to write a message to the connection
	pongWait       = 60 * time.Second    // time allowed to read the next pong message from the connection
	pingPeriod     = (pongWait * 9) / 10 // time interval for sending ping message to the connection
	maxMessageSize = 512                 // maxmimum message size allowed from connection
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	hub     *WsHub
	user    User
	conn    *websocket.Conn
	message chan Message
}

const bufferSize int = 100

type WsHub struct {
	broker     MessageBroker
	clients    sync.Map
	register   chan *Client
	unregister chan *Client
}

func NewWsHub(broker MessageBroker) *WsHub {
	return &WsHub{
		broker:     broker,
		register:   make(chan *Client, bufferSize),
		unregister: make(chan *Client, bufferSize),
	}
}

func (h *WsHub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case client := <-h.register:
			h.addClient(client)
		case client := <-h.unregister:
			h.removeClient(client)
		}
	}
}

func (h *WsHub) addClient(client *Client) {
	var clients []*Client

	userClients, ok := h.clients.Load(client.user.ID)
	if ok {
		clients = userClients.([]*Client)
	}

	clients = append(clients, client)

	h.clients.Store(client.user.ID, clients)
}

func (h *WsHub) removeClient(client *Client) {
	userClients, ok := h.clients.Load(client.user.ID)
	if !ok {
		return
	}

	clients := userClients.([]*Client)

	clients = slices.DeleteFunc(clients, func(c *Client) bool {
		return c == client
	})

	if len(clients) == 0 {
		h.clients.Delete(client.user.ID)
	} else {
		h.clients.Store(client.user.ID, clients)
	}

	close(client.message)
}

func (h *WsHub) serveWebSockets(user User, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Error().Msg(err.Error())
		return
	}

	client := &Client{
		hub:     h,
		user:    user,
		conn:    conn,
		message: make(chan Message, bufferSize),
	}

	client.hub.register <- client

	go client.read()
	go client.write()
}

func (h *WsHub) handleMessageFromBroker(topic MessageBrokerTopic, msg []byte) error {
	switch topic {
	case MessageCreatedTopic:
		{
			var payload Message
			if err := json.Unmarshal(msg, &payload); err != nil {
				log.Error().Msg(err.Error())
				return err
			}

			if clients, ok := h.clients.Load(payload.RecipientID); ok {
				for _, client := range clients.([]*Client) {
					go func(client *Client) {
						client.message <- payload
					}(client)
				}
			}
		}
	}

	return nil
}

// reads pong message and JSON payload from websocket connection
func (c *Client) read() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, payload, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Error().Msg(err.Error())
			}

			break
		}

		err = c.hub.broker.publish(MessageCreatedTopic, payload)
		if err != nil {
			log.Error().Msg(err.Error())
		}
	}
}

// writes ping message and JSON payload on websocket connection
// closes connection when client is unresponsive
func (c *Client) write() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.message:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteJSON(message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

package mig

import (
	"context"
	"encoding/json"
	"net/http"
	"slices"
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
	hub     *Hub
	user    User
	conn    *websocket.Conn
	message chan Message
}

type Hub struct {
	broker     MessageBroker
	clients    map[int64][]*Client
	register   chan *Client
	unregister chan *Client
}

func NewHub(broker MessageBroker) *Hub {
	return &Hub{
		broker:     broker,
		clients:    make(map[int64][]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case client := <-h.register:
			h.clients[client.user.ID] = append(h.clients[client.user.ID], client)
		case client := <-h.unregister:
			if _, ok := h.clients[client.user.ID]; ok {
				h.clients[client.user.ID] = slices.DeleteFunc(h.clients[client.user.ID], func(c *Client) bool {
					return client == c
				})

				if len(h.clients[client.user.ID]) == 0 {
					delete(h.clients, client.user.ID)
				}

				close(client.message)
			}
		}
	}
}

func (h *Hub) ServeWebSockets(user User, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Error().Msg(err.Error())
		return
	}

	client := &Client{
		hub:     h,
		user:    user,
		conn:    conn,
		message: make(chan Message),
	}

	client.hub.register <- client

	go client.read()
	go client.write()
}

func handleKafkaMsgReceived(msg []byte) error {
	var payload Message

	err := json.Unmarshal(msg, &payload)

	return err
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
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Error().Msg(err.Error())
			}

			break
		}

		var payload Message
		if err := json.Unmarshal(msg, &payload); err != nil {
			log.Error().Msg(err.Error())

			c.conn.WriteMessage(websocket.TextMessage, []byte(`{"code":"00001","message":"JSON failed, please contact IT."}`))

			continue
		}

		c.hub.broker.publish("mig.messages.created", payload)
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

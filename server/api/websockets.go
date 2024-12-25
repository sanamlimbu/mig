package api

import (
	"context"
	"encoding/json"
	"fmt"
	"mig"
	"mig/auth"
	"mig/messagebroker"
	"mig/user"
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

// Reference - https://github.com/gorilla/websocket/blob/main/examples/chat/Client.go
// Reference - https://devcenter.heroku.com/articles/websocket-security

const (
	// time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// send pings to peer with this period, must be less than pongWait
	pingPeriod = (pongWait * 9) / 10

	// maximum message size allowed from peer
	maxMessageSize = 1024
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	hub  *WsHub
	user *mig.User
	conn *websocket.Conn
	send chan []byte // buffered channel for outbound message
}

const registerBufferSize int = 100
const messageBufferSize int = 256

type WsHub struct {
	broker      messagebroker.MessageBroker
	clients     sync.Map
	register    chan *Client
	unregister  chan *Client
	authService *auth.Service
	userService *user.Service
}

func NewWsHub(broker messagebroker.MessageBroker, authSerivce *auth.Service, userSerivce *user.Service) (*WsHub, error) {
	if broker == nil {
		return nil, fmt.Errorf("missing message broker")
	}

	if authSerivce == nil {
		return nil, fmt.Errorf("missing auth service")
	}

	if userSerivce == nil {
		return nil, fmt.Errorf("missing user service")
	}

	hub := &WsHub{
		broker:      broker,
		register:    make(chan *Client, registerBufferSize),
		unregister:  make(chan *Client, registerBufferSize),
		authService: authSerivce,
		userService: userSerivce,
	}

	return hub, nil
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

	if client.user == nil {
		return
	}

	userClients, ok := h.clients.Load(client.user.ID)
	if ok {
		clients = userClients.([]*Client)
	}

	clients = append(clients, client)

	h.clients.Store(client.user.ID, clients)
}

func (h *WsHub) removeClient(client *Client) {
	if client.user == nil {
		return
	}

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

	close(client.send)
}

func (h *WsHub) serveWebSockets(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Error().Msg(err.Error())
		return
	}

	client := &Client{
		hub:  h,
		user: nil,
		conn: conn,
		send: make(chan []byte, messageBufferSize),
	}

	go client.read()
	go client.write()
}

func (h *WsHub) HandleBrokerMessage(topic messagebroker.Topic, msg []byte) error {
	switch topic {
	case messagebroker.TopicMessageCreated:
		{
			var payload messagebroker.TopicMessageCreatedPayload

			if err := json.Unmarshal(msg, &payload); err != nil {
				log.Error().Msg(err.Error())
				return err
			}

			if clients, ok := h.clients.Load(payload.RecipientID); ok {
				for _, client := range clients.([]*Client) {
					go func(client *Client) {
						client.send <- msg
					}(client)
				}
			}
		}
	case messagebroker.TopicMessageUpdated:
		{
			var payload messagebroker.TopicMessageUpdatedPayload

			if err := json.Unmarshal(msg, &payload); err != nil {
				log.Error().Msg(err.Error())
				return err
			}

			if clients, ok := h.clients.Load(payload.RecipientID); ok {
				for _, client := range clients.([]*Client) {
					go func(client *Client) {
						client.send <- msg
					}(client)
				}
			}
		}
	case messagebroker.TopicMessageDeleted:
		{
			var payload messagebroker.TopicMessageDeletedPayload

			if err := json.Unmarshal(msg, &payload); err != nil {
				log.Error().Msg(err.Error())
				return err
			}

			if clients, ok := h.clients.Load(payload.RecipientID); ok {
				for _, client := range clients.([]*Client) {
					go func(client *Client) {
						client.send <- msg
					}(client)
				}
			}
		}
	default:
		return fmt.Errorf("invalid message broker topic: %s", topic)
	}

	return nil
}

// read pongs message from websocket connection
func (c *Client) read() {
	defer func() {
		c.hub.unregister <- c
		if err := c.conn.Close(); err != nil {
			log.Error().Msg(err.Error())
		}
	}()

	c.conn.SetReadLimit(maxMessageSize)

	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Error().Msg(err.Error())
	}

	c.conn.SetPongHandler(func(string) error {
		if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			log.Error().Msg(err.Error())
		}
		return nil
	})

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Error().Msg(err.Error())
			}
			break
		}

		if c.user == nil {
			if err := c.register(data); err != nil {
				log.Error().Msg(err.Error())
				break
			}
		}

		payload, err := c.parseMessage(data)
		if err != nil {
			log.Error().Msg(err.Error())
			continue
		}

		bytes, err := json.Marshal(payload)
		if err != nil {
			log.Error().Msg(err.Error())
			continue
		}

		err = c.hub.broker.Publish(payload.GetTopic(), bytes)
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
		if err := c.conn.Close(); err != nil {
			log.Error().Msg(err.Error())
		}
	}()

	for {
		select {
		case message, ok := <-c.send:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				log.Error().Msg(err.Error())
			}

			if !ok {
				if err := c.conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					log.Error().Msg(err.Error())
				}
				return
			}

			if err := c.conn.WriteJSON(message); err != nil {
				return
			}

		case <-ticker.C:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				log.Error().Msg(err.Error())
			}
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

type WebsocketAuthMessage struct {
	Token string `json:"token"`
}

// register parses auth message and registers client to websocket hub.
func (c *Client) register(data []byte) error {
	var msg WebsocketAuthMessage

	if err := json.Unmarshal(data, &msg); err != nil {
		return err
	}

	userID, err := c.hub.authService.ParseJwtToken(msg.Token)
	if err != nil {
		return err
	}

	user, err := c.hub.userService.GetUser(context.Background(), userID)
	if err != nil {
		return err
	}

	c.user = &user

	c.hub.register <- c

	return nil
}

type WebsocketMessageType string

const (
	WebsocketMessageTypeMessageCreated WebsocketMessageType = "message_created"
	WebsocketMessageTypeMessageUpdated WebsocketMessageType = "message_updated"
	WebsocketMessageTypeMessageDeleted WebsocketMessageType = "message_deleted"
)

type WebsocketMessage struct {
	MessageType WebsocketMessageType `json:"message_type"`
	Payload     any                  `json:"payload"`
}

// parseMessage parses data and returns message broker's message.
func (c *Client) parseMessage(data []byte) (messagebroker.Message, error) {
	var msg WebsocketMessage

	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("unable to unmarshal websocket message: %w", err)
	}

	bytes, err := json.Marshal(msg.Payload)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal websocket message payload: %w", err)
	}

	switch msg.MessageType {
	case WebsocketMessageTypeMessageCreated:
		var payload messagebroker.TopicMessageCreatedPayload
		if err := json.Unmarshal(bytes, &payload); err != nil {
			return nil, fmt.Errorf("unable to unmarshal message created payload: %w", err)
		}
		return payload, nil

	case WebsocketMessageTypeMessageUpdated:
		var payload messagebroker.TopicMessageUpdatedPayload
		if err := json.Unmarshal(bytes, &payload); err != nil {
			return nil, fmt.Errorf("unable to unmarshal message updated payload: %w", err)
		}
		return payload, nil

	case WebsocketMessageTypeMessageDeleted:
		var payload messagebroker.TopicMessageDeletedPayload
		if err := json.Unmarshal(bytes, &payload); err != nil {
			return nil, fmt.Errorf("unable to unmarshal message deleted payload: %w", err)
		}
		return payload, nil

	default:
		return nil, fmt.Errorf("unknown websocket message type: %s", msg.MessageType)
	}
}

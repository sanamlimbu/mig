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

type WebsocketMessageType string

const (
	MessageCreatedWebsocketMessageType WebsocketMessageType = "message_created"
	MessageUpdatedWebsocketMessageType WebsocketMessageType = "message_updated"
	MessageDeletedWebsocketMessageType WebsocketMessageType = "message_deleted"
	AuthenticationWebsocketMessageType WebsocketMessageType = "authentication"
)

type WebsocketMessage struct {
	Type    WebsocketMessageType `json:"type"`
	Payload any                  `json:"payload"`
}

func (h *WsHub) HandleBrokerMessage(topic messagebroker.Topic, msg []byte) error {
	switch topic {
	case messagebroker.MessageCreatedTopic:
		{
			var payload messagebroker.MessageCreatedTopicPayload
			if err := json.Unmarshal(msg, &payload); err != nil {
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
	case messagebroker.MessageUpdatedTopic:
		{
			var payload messagebroker.MessageUpdatedTopicPayload
			if err := json.Unmarshal(msg, &payload); err != nil {
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
	case messagebroker.MessageDeletedTopic:
		{
			var payload messagebroker.MessageDeletedTopicPayload
			if err := json.Unmarshal(msg, &payload); err != nil {
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
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

loop:
	for {

		var msg WebsocketMessage
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Error().Msg(err.Error())
			}
			break loop
		}

		data, err := json.Marshal(msg.Payload)
		if err != nil {
			log.Error().Msg(err.Error())
			break loop
		}

		switch msg.Type {
		case AuthenticationWebsocketMessageType:
			{
				if c.user == nil {
					if err := c.register(data); err != nil {
						log.Error().Msg(err.Error())
						break loop
					}
				}
			}
		case MessageCreatedWebsocketMessageType, MessageUpdatedWebsocketMessageType, MessageDeletedWebsocketMessageType:
			{
				payload, err := c.parseBrokerMessage(msg.Type, data)
				if err != nil {
					log.Error().Msg(err.Error())
					break loop
				}

				err = c.hub.broker.Publish(payload.GetTopic(), data)
				if err != nil {
					log.Error().Msg(err.Error())
				}
			}
		default:
			{
				log.Error().Msg(fmt.Sprintf("invalid websocket message type: %s", msg.Type))
				break loop
			}
		}
	}
}

// writes ping message and JSON payload on websocket connection
// closes connection when client is unresponsive
func (c *Client) write() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			{
				_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
				if !ok {
					_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}

				if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
					return
				}
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// register parses auth message and registers client to websocket hub.
func (c *Client) register(data []byte) error {
	var msg struct {
		AccessToken string `json:"access_token"`
	}

	if err := json.Unmarshal(data, &msg); err != nil {
		return err
	}

	valid, claims, err := c.hub.authService.VerifyAccessToken(msg.AccessToken)
	if err != nil {
		return err
	}

	if !valid {
		return fmt.Errorf("invalid access token")
	}

	user, err := c.hub.userService.GetUser(context.Background(), claims.UserID)
	if err != nil {
		return err
	}

	c.user = &user

	c.hub.register <- c

	return nil
}

func (c *Client) parseBrokerMessage(msgType WebsocketMessageType, data []byte) (messagebroker.Message, error) {
	switch msgType {
	case MessageCreatedWebsocketMessageType:
		var payload messagebroker.MessageCreatedTopicPayload
		if err := json.Unmarshal(data, &payload); err != nil {
			return nil, fmt.Errorf("unable to unmarshal message created payload: %w", err)
		}
		return payload, nil

	case MessageUpdatedWebsocketMessageType:
		var payload messagebroker.MessageUpdatedTopicPayload
		if err := json.Unmarshal(data, &payload); err != nil {
			return nil, fmt.Errorf("unable to unmarshal message updated payload: %w", err)
		}
		return payload, nil

	case MessageDeletedWebsocketMessageType:
		var payload messagebroker.MessageDeletedTopicPayload
		if err := json.Unmarshal(data, &payload); err != nil {
			return nil, fmt.Errorf("unable to unmarshal message deleted payload: %w", err)
		}
		return payload, nil

	default:
		return nil, fmt.Errorf("unknown websocket message type: %s", msgType)
	}
}

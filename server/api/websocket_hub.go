package api

import (
	"context"
	"encoding/json"
	"fmt"
	"mig/auth"
	"mig/chatroom"
	"mig/messagebroker"
	"mig/user"
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
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

const registerBufferSize int = 100
const messageBufferSize int = 256

type WsHub struct {
	broker          messagebroker.MessageBroker
	clients         sync.Map
	register        chan *Client
	unregister      chan *Client
	chatrooms       sync.Map
	authService     *auth.Service
	userService     *user.Service
	chatroomService *chatroom.Service
}

func NewWsHub(broker messagebroker.MessageBroker, authSerivce *auth.Service, userSerivce *user.Service, chatroomService *chatroom.Service) (*WsHub, error) {
	if broker == nil {
		return nil, fmt.Errorf("missing message broker")
	}

	if authSerivce == nil {
		return nil, fmt.Errorf("missing auth service")
	}

	if userSerivce == nil {
		return nil, fmt.Errorf("missing user service")
	}

	if chatroomService == nil {
		return nil, fmt.Errorf("missing chatroom service")
	}

	hub := &WsHub{
		broker:          broker,
		register:        make(chan *Client, registerBufferSize),
		unregister:      make(chan *Client, registerBufferSize),
		authService:     authSerivce,
		userService:     userSerivce,
		chatroomService: chatroomService,
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
	if client.user == nil {
		return
	}

	var clients []*Client

	if client.clientType == clientTypeChatroom {
		existing, ok := h.chatrooms.Load(client.chatroom.ID)
		if ok {
			clients = existing.([]*Client)
		}

		clients = append(clients, client)
		h.chatrooms.Store(client.chatroom.ID, clients)
	} else {
		existing, ok := h.clients.Load(client.user.ID)
		if ok {
			clients = existing.([]*Client)
		}

		clients = append(clients, client)
		h.clients.Store(client.user.ID, clients)
	}
}

func (h *WsHub) removeClient(client *Client) {
	if client.user == nil {
		return
	}

	if client.clientType == clientTypeChatroom {
		existing, ok := h.chatrooms.Load(client.chatroom.ID)
		if !ok {
			return
		}

		clients := existing.([]*Client)

		clients = slices.DeleteFunc(clients, func(c *Client) bool {
			return c == client
		})

		if len(clients) == 0 {
			h.chatrooms.Delete(client.chatroom.ID)
		} else {
			h.chatrooms.Store(client.chatroom.ID, clients)
		}

	} else {
		existing, ok := h.clients.Load(client.user.ID)
		if !ok {
			return
		}

		clients := existing.([]*Client)

		clients = slices.DeleteFunc(clients, func(c *Client) bool {
			return c == client
		})

		if len(clients) == 0 {
			h.clients.Delete(client.user.ID)
		} else {
			h.clients.Store(client.user.ID, clients)
		}
	}

	close(client.send)
}

func (h *WsHub) serveUserWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Msg(err.Error())
		return
	}

	client := &Client{
		hub:        h,
		conn:       conn,
		send:       make(chan []byte, messageBufferSize),
		clientType: clientTypeUser,
	}

	go client.read()
	go client.write()
}

func (h *WsHub) serveChatroomWebSocket(w http.ResponseWriter, r *http.Request) {
	chatroomID := chi.URLParam(r, "chatroom_id")
	chatroom, err := h.chatroomService.GetChatroomWithCreator(r.Context(), chatroomID)
	if err != nil {
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Msg(err.Error())
		return
	}

	client := &Client{
		hub:        h,
		chatroom:   &chatroom,
		conn:       conn,
		send:       make(chan []byte, messageBufferSize),
		clientType: clientTypeChatroom,
	}

	go client.read()
	go client.write()
}

type WebsocketMessageType string

const (
	WebsocketMessageTypeMessageCreated WebsocketMessageType = "message_created"
	WebsocketMessageTypeMessageUpdated WebsocketMessageType = "message_updated"
	WebsocketMessageTypeMessageDeleted WebsocketMessageType = "message_deleted"
	WebsocketMessageTypeAuthentication WebsocketMessageType = "authentication"
)

type WebsocketMessage struct {
	Type    WebsocketMessageType `json:"type"`
	Payload any                  `json:"payload"`
}

func (h *WsHub) HandleBrokerMessage(topic messagebroker.Topic, msg []byte) error {
	switch topic {
	case messagebroker.TopicMessageCreated:
		{
			var payload messagebroker.MessageCreatedTopicPayload
			if err := json.Unmarshal(msg, &payload); err != nil {
				return err
			}

			send, err := json.Marshal(WebsocketMessage{
				Type:    WebsocketMessageTypeMessageCreated,
				Payload: payload,
			})
			if err != nil {
				return err
			}

			if payload.Type == "chatroom" {
				if clients, ok := h.chatrooms.Load(payload.RecipientID); ok {
					for _, client := range clients.([]*Client) {
						go func(client *Client) {
							client.send <- send
						}(client)
					}
				}
			} else {
				if clients, ok := h.clients.Load(payload.RecipientID); ok {
					for _, client := range clients.([]*Client) {
						go func(client *Client) {
							client.send <- send
						}(client)
					}
				}
			}
		}
	case messagebroker.TopicMessageUpdated:
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
	case messagebroker.TopicMessageDeleted:
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

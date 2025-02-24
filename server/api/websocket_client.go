package api

import (
	"context"
	"encoding/json"
	"fmt"
	"mig"
	"mig/messagebroker"
	"mig/user"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

type Client struct {
	hub        *WsHub
	user       *mig.User
	chatroom   *mig.Chatroom
	conn       *websocket.Conn
	send       chan []byte
	clientType clientType
}

type chatroomRegister struct {
	client     *Client
	chatroomID string
}

type chatroomUnregister struct {
	client     *Client
	chatroomID string
}

type clientType string

const (
	clientTypeUser     clientType = "user"
	clientTypeChatroom clientType = "chatroom"
)

// read pongs message from websocket connection
func (c *Client) read() {
	defer func() {
		if c.clientType == clientTypeChatroom {
			c.hub.chatroomUnregister <- chatroomUnregister{
				client:     c,
				chatroomID: c.chatroom.ID,
			}
			_ = c.conn.Close()
		} else {
			c.hub.unregister <- c
			_ = c.conn.Close()
		}
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
		case WebsocketMessageTypeAuthentication:
			{
				if c.user == nil {
					if err := c.register(data); err != nil {
						log.Error().Msg(err.Error())
						break loop
					}
				}
			}
		case WebsocketMessageTypeMessageCreated, WebsocketMessageTypeMessageUpdated, WebsocketMessageTypeMessageDeleted:
			{
				payload, err := c.parseBrokerMessage(msg.Type, data)
				if err != nil {
					log.Error().Msg(err.Error())
					break loop
				}

				if msg.Type == WebsocketMessageTypeMessageCreated {
					if payload, ok := payload.(messagebroker.MessageCreatedTopicPayload); ok {
						msg, err := c.hub.userService.SaveMessage(context.Background(), user.SaveMessageParams{
							ID:          payload.ID,
							SenderID:    payload.SenderID,
							RecipientID: payload.RecipientID,
							Content:     payload.Content,
							Type:        payload.Type,
						})
						if err != nil {
							log.Error().Msg(err.Error())
							break loop
						}

						payload.CreatedAt = msg.CreatedAt

						data, err := json.Marshal(payload)
						if err != nil {
							log.Error().Msg(err.Error())
							break loop
						}

						err = c.hub.broker.Publish(messagebroker.TopicMessageCreated, data)
						if err != nil {
							log.Error().Msg(err.Error())
						}
					} else {
						log.Error().Msg("unexpected payload type for message.created topic")
						break loop
					}

				} else if msg.Type == WebsocketMessageTypeMessageUpdated {
					if payload, ok := payload.(messagebroker.MessageUpdatedTopicPayload); ok {
						msg, err := c.hub.userService.UpdateMessage(context.Background(), user.UpdateMessageParams{
							Content:       payload.Content,
							WorkflowState: payload.WorflowState,
							IsRead:        payload.IsRead,
						})
						if err != nil {
							log.Error().Msg(err.Error())
							break loop
						}

						payload.CreatedAt = msg.CreatedAt
						payload.UpdatedAt = msg.UpdatedAt

						data, err := json.Marshal(payload)
						if err != nil {
							log.Error().Msg(err.Error())
							break loop
						}

						err = c.hub.broker.Publish(messagebroker.TopicMessageCreated, data)
						if err != nil {
							log.Error().Msg(err.Error())
						}
					} else {
						log.Error().Msg("unexpected payload type for message.updated topic")
						break loop
					}

				} else if msg.Type == WebsocketMessageTypeMessageDeleted {
					if payload, ok := payload.(messagebroker.MessageDeletedTopicPayload); ok {
						if err := c.hub.userService.DeleteMessage(context.Background(), payload.ID); err != nil {
							log.Error().Msg(err.Error())
							break loop
						}

						err := c.hub.broker.Publish(payload.GetTopic(), data)
						if err != nil {
							log.Error().Msg(err.Error())
						}
					} else {
						log.Error().Msg("unexpected payload type for message.deleted topic")
						break loop
					}
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

	claims, err := c.hub.authService.VerifyAccessToken(msg.AccessToken)
	if err != nil {
		return err
	}

	user, err := c.hub.userService.GetUser(context.Background(), claims.UserID)
	if err != nil {
		return err
	}

	c.user = &user

	if c.clientType == clientTypeChatroom {
		c.hub.chatroomRegister <- chatroomRegister{
			client:     c,
			chatroomID: c.chatroom.ID,
		}
	} else {
		c.hub.register <- c
	}

	return nil
}

func (c *Client) parseBrokerMessage(msgType WebsocketMessageType, data []byte) (messagebroker.Message, error) {
	switch msgType {
	case WebsocketMessageTypeMessageCreated:
		var payload messagebroker.MessageCreatedTopicPayload
		if err := json.Unmarshal(data, &payload); err != nil {
			return nil, fmt.Errorf("unable to unmarshal message created payload: %w", err)
		}
		return payload, nil

	case WebsocketMessageTypeMessageUpdated:
		var payload messagebroker.MessageUpdatedTopicPayload
		if err := json.Unmarshal(data, &payload); err != nil {
			return nil, fmt.Errorf("unable to unmarshal message updated payload: %w", err)
		}
		return payload, nil

	case WebsocketMessageTypeMessageDeleted:
		var payload messagebroker.MessageDeletedTopicPayload
		if err := json.Unmarshal(data, &payload); err != nil {
			return nil, fmt.Errorf("unable to unmarshal message deleted payload: %w", err)
		}
		return payload, nil

	default:
		return nil, fmt.Errorf("unknown websocket message type: %s", msgType)
	}
}

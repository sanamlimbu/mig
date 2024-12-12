package messagebroker

import (
	"fmt"
	"mig"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

type MessageBroker interface {
	Publish(topic Topic, message []byte) error
	Subscribe(topic Topic, msgHandler IncommingMessageHandler)
	Close()
}

type Topic string

const (
	MessageCreatedTopic Topic = "message.created"
	MessageDeletedTopic Topic = "message.deleted"
	MessageUpdatedTopic Topic = "message.updated"
)

type Message interface {
	GetTopic() string
}

type MessageCreatedTopicMessage struct {
	ID          int64           `json:"id"`
	SenderID    int64           `json:"sender_id"`
	RecipientID int64           `json:"recipient_id"` // user id or chatroom id
	Content     string          `json:"content"`
	MessageType mig.MessageType `json:"message_type"`
}

func (m MessageCreatedTopicMessage) GetTopic() string {
	return string(MessageCreatedTopic)
}

type MessageDeletedTopicMessage struct {
	ID          int64           `json:"id"`
	SenderID    int64           `json:"sender_id"`
	RecipientID int64           `json:"recipient_id"` // user id or chatroom id
	Content     string          `json:"content"`
	MessageType mig.MessageType `json:"message_type"`
}

func (m MessageDeletedTopicMessage) GetTopic() string {
	return string(MessageDeletedTopic)
}

type IncommingMessageHandler interface {
	HandleBrokerMessage(topic Topic, msg []byte) error
}

type Nats struct {
	conn *nats.Conn
}

func NewNats(url string) (*Nats, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}

	nats := &Nats{
		conn: nc,
	}

	return nats, nil
}

func (n *Nats) Publish(topic Topic, message []byte) error {
	err := n.conn.Publish(string(topic), message)
	if err != nil {
		return err
	}

	return nil
}

func (n *Nats) Subscribe(topic Topic, msgHandler IncommingMessageHandler) {
	n.conn.Subscribe(string(topic), func(msg *nats.Msg) {
		msgHandler.HandleBrokerMessage(topic, msg.Data)
	})
	n.conn.Flush()

	if err := n.conn.LastError(); err != nil {
		msg := fmt.Sprintf("subscribed topic %s: %s", string(topic), err.Error())
		log.Error().Msg(msg)
	}
}

func (n *Nats) Close() {
	log.Info().Msg("closing NATS connection...")
	n.conn.Close()
}

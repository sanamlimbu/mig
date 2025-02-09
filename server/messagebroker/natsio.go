package messagebroker

import (
	"fmt"
	"mig"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

type MessageBroker interface {
	Publish(topic Topic, message []byte) error
	Subscribe(topic Topic, msgHandler IncommingMessageHandler) error
	Close()
}

type Topic string

const (
	TopicMessageCreated Topic = "message.created"
	TopicMessageDeleted Topic = "message.deleted"
	TopicMessageUpdated Topic = "message.updated"
)

func GetAllTopics() []Topic {
	return []Topic{
		TopicMessageCreated,
		TopicMessageUpdated,
		TopicMessageDeleted,
	}
}

type Message interface {
	GetTopic() Topic
}

type MessageCreatedTopicPayload struct {
	ID          string          `json:"id"`
	SenderID    string          `json:"sender_id"`
	RecipientID string          `json:"recipient_id"`
	Content     string          `json:"content"`
	MessageType mig.MessageType `json:"message_type"`
}

func (m MessageCreatedTopicPayload) GetTopic() Topic {
	return TopicMessageCreated
}

type MessageUpdatedTopicPayload struct {
	ID          string          `json:"id"`
	SenderID    string          `json:"sender_id"`
	RecipientID string          `json:"recipient_id"`
	Content     string          `json:"content"`
	MessageType mig.MessageType `json:"message_type"`
}

func (m MessageUpdatedTopicPayload) GetTopic() Topic {
	return TopicMessageUpdated
}

type MessageDeletedTopicPayload struct {
	ID          string          `json:"id"`
	SenderID    string          `json:"sender_id"`
	RecipientID string          `json:"recipient_id"`
	Content     string          `json:"content"`
	MessageType mig.MessageType `json:"message_type"`
}

func (m MessageDeletedTopicPayload) GetTopic() Topic {
	return TopicMessageDeleted
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

func (n *Nats) Subscribe(topic Topic, msgHandler IncommingMessageHandler) error {
	_, err := n.conn.Subscribe(string(topic), func(msg *nats.Msg) {
		if err := msgHandler.HandleBrokerMessage(topic, msg.Data); err != nil {
			log.Error().Msg(err.Error())
		}
	})

	if err != nil {
		return err
	}

	if err := n.conn.Flush(); err != nil {
		log.Error().Msg(err.Error())
	}

	if err := n.conn.LastError(); err != nil {
		msg := fmt.Sprintf("subscribed topic %s: %s", string(topic), err.Error())
		log.Error().Msg(msg)
	}

	return nil
}

func (n *Nats) Close() {
	log.Info().Msg("closing NATS connection...")
	n.conn.Close()
}

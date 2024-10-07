package mig

type MessageBroker interface {
	publish(topic MessageBrokerTopic, message []byte) error
	Subscribe(topic MessageBrokerTopic, msgHandler BrokerMessageHandler)
	Close()
}

type MessageType string

const (
	MessageTypePrivate  MessageType = "private"
	MessageTypeChatroom MessageType = "chatroom"
)

type Message struct {
	ID          int64       `json:"id"`
	SenderID    int64       `json:"sender_id"`
	RecipientID int64       `json:"recipient_id"` // user id or chatroom id
	Content     string      `json:"content"`
	MessageType MessageType `json:"message_type"`
}

type MessageBrokerTopic string

const (
	MessageCreatedTopic MessageBrokerTopic = "message.created"
)

type BrokerMessageHandler interface {
	handleMessageFromBroker(topic MessageBrokerTopic, msg []byte) error
}

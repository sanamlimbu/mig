package mig

type MessageBroker interface {
	publish(topic string, message Message) error
}

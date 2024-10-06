package mig

import (
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

// Nats implement MessageBroker interface
type Nats struct {
	conn *nats.Conn
}

func NewNats(url string) (*Nats, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}

	nats := Nats{
		conn: nc,
	}

	return &nats, nil
}

func (n *Nats) publish(subject MessageBrokerTopic, message []byte) error {
	err := n.conn.Publish(string(subject), message)
	if err != nil {
		msg := fmt.Sprintf("publish: %s", err.Error())
		log.Error().Msg(msg)
		return err
	}

	return nil
}

func (n *Nats) Subscribe(subject MessageBrokerTopic, msgHandler BrokerMessageHandler) {
	n.conn.Subscribe(string(subject), func(msg *nats.Msg) {
		msgHandler.handleMessageFromBroker(subject, msg.Data)
	})
	n.conn.Flush()

	if err := n.conn.LastError(); err != nil {
		msg := fmt.Sprintf("subscribe: %s", err.Error())
		log.Error().Msg(msg)
	}
}

func (n *Nats) Close() {
	log.Info().Msg("closing NATS connection...")
	n.conn.Close()
}

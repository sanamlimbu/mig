package mig

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

type Nats struct {
	conn *nats.Conn
}

func NewNats(protocol, user, pass, host, port string) (*Nats, error) {
	url := fmt.Sprintf("%s://%s:%s@%s:%s", protocol, user, pass, host, port)

	nc, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}

	nats := Nats{
		conn: nc,
	}

	return &nats, nil
}

func (n *Nats) publish(subject string, message Message) error {
	var data bytes.Buffer

	enc := gob.NewEncoder(&data)

	err := enc.Encode(message)
	if err != nil {
		msg := fmt.Sprintf("encode: %s", err.Error())
		log.Error().Msg(msg)
		return err
	}

	err = n.conn.Publish(subject, data.Bytes())
	if err != nil {
		msg := fmt.Sprintf("publish: %s", err.Error())
		log.Error().Msg(msg)
		return err
	}

	return nil
}

func (n *Nats) subscribe(subject string) {
	n.conn.Subscribe(subject, func(msg *nats.Msg) {
		handleMessage(msg)
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

func handleMessage(msg *nats.Msg) {

}

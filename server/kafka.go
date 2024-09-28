package mig

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/rs/zerolog/log"
)

type MessageType string

const (
	MessageTypePrivate MessageType = "private"
	MessageTypeGroup   MessageType = "group"
)

type Message struct {
	ID          int64       `json:"id"`
	SenderID    int64       `json:"sender_id"`
	RecipientID int64       `json:"recipient_id"` // user id or group id
	Content     string      `json:"content"`
	MessageType MessageType `json:"message_type"`
}

type Kafka struct {
	producer sarama.SyncProducer
	consumer sarama.ConsumerGroup
	topics   []string
}

func NewKafka(brokers []string, version sarama.KafkaVersion, topics []string, assignor, group string) (*Kafka, error) {
	config := sarama.NewConfig()
	config.Version = version
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.Retry.Max = 10
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Successes = true

	switch assignor {
	case "sticky":
		config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategySticky()}
	case "roundrobin":
		config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	case "range":
		config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRange()}
	default:
		return nil, fmt.Errorf("unrecognized consumer group partition assignor: %s", assignor)
	}

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	consumer, err := sarama.NewConsumerGroup(brokers, group, config)
	if err != nil {
		return nil, err
	}

	kafka := &Kafka{
		producer: producer,
		consumer: consumer,
		topics:   topics,
	}

	return kafka, nil
}

func (k *Kafka) Close() error {
	return k.consumer.Close()
}

func (k *Kafka) SendMessage(m Message, topic string) error {
	payload, err := json.Marshal(m)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(payload),
	}
	_, _, err = k.producer.SendMessage(msg)

	return err
}

func (k *Kafka) Consume(ctx context.Context, c *Consumer) {
	for {
		if err := k.consumer.Consume(ctx, k.topics, c); err != nil {
			if errors.Is(err, sarama.ErrClosedConsumerGroup) {
				return
			}
			log.Error().Msg(err.Error())
		}
	}
}

// Reference - https://github.com/IBM/sarama/blob/main/examples/consumergroup/main.go

// Consumer represents a Sarama consumer group consumer
type Consumer struct {
	Ready chan bool
	hub   *Hub
}

func NewConsumer(hub *Hub) *Consumer {
	return &Consumer{
		Ready: make(chan bool),
		hub:   hub,
	}
}

func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	close(consumer.Ready)

	return nil
}

func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case msg, ok := <-claim.Messages():
			if !ok {
				log.Error().Msg("Kafka message channel was closed")
				return nil
			}

			var payload Message
			if err := json.Unmarshal(msg.Value, &payload); err != nil {
				log.Error().Msg(err.Error())
				break
			}

			for _, client := range consumer.hub.clients[payload.RecipientID] {
				client.message <- payload
			}

			session.MarkMessage(msg, "")
		case <-session.Context().Done():
			return nil
		}
	}
}

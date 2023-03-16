package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/Shopify/sarama"
)

type Sender interface {
	Send(msg []byte, key fmt.Stringer, ctx context.Context) error
}

type KafkaSender struct {
	Dest Destination
	Producer sarama.AsyncProducer
}

var ErrEmptyDestination = errors.New("empty destination")

func (k *KafkaSender) Send(msg []byte, key fmt.Stringer, ctx context.Context) error {
	topic := k.Dest.GetDestination(nil)
	if topic == "" {
		return ErrEmptyDestination
	}

	if key == nil {
		k.Producer.Input() <- &sarama.ProducerMessage{
			Topic: topic,
			Value: sarama.ByteEncoder(msg),
		}
	} else {
		k.Producer.Input() <- &sarama.ProducerMessage{
			Topic: topic,
			Key: sarama.StringEncoder(key.String()),
			Value: sarama.ByteEncoder(msg),
		}
	}

	return nil
}
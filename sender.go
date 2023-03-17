package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/Shopify/sarama"
)

type Sender interface {
	Send(ctx context.Context, msg []byte, key fmt.Stringer) error
}

type KafkaSender struct {
	Dest     Destination
	Producer sarama.AsyncProducer
}

var ErrEmptyDestination = errors.New("empty destination")

func (k *KafkaSender) Send(ctx context.Context, msg []byte, key fmt.Stringer) error {
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
			Key:   sarama.StringEncoder(key.String()),
			Value: sarama.ByteEncoder(msg),
		}
	}

	return nil
}

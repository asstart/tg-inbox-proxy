package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/Shopify/sarama"
	"github.com/asstart/tg-inbox-proxy/message"
	"google.golang.org/protobuf/proto"
)

type Sender interface {
	Send(ctx context.Context, msg []byte, key fmt.Stringer) error
	Close() error
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

func (k *KafkaSender) Close() error {
	return k.Producer.Close()
}

type FileSender struct {
	File *os.File
}

func (s *FileSender) Send(ctx context.Context, msg []byte, key fmt.Stringer) error {
	var strMsg message.Message
	if err := proto.Unmarshal(msg, &strMsg); err != nil {
		log.Printf("ERROR: %s\n", err)
	}

	if key == nil {
		fmt.Fprintf(s.File, "%v\n", &strMsg)
	} else {
		fmt.Fprintf(s.File, "%s: %v\n", key.String(), &strMsg)
	}

	return nil
}

func (s *FileSender) Close() error {
	return s.File.Close()
}

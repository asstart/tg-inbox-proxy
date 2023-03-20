package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/asstart/tg-inbox-proxy/message"
	"google.golang.org/protobuf/proto"
	tele "gopkg.in/telebot.v3"
)

type Handler interface {
	Handle(ctx context.Context, msg tele.Context) ([]byte, fmt.Stringer, error)
}

type TextMessageHandler struct{}

func (h *TextMessageHandler) Handle(ctx context.Context, msg tele.Context) ([]byte, fmt.Stringer, error) {
	err := validate(msg)
	if err != nil {
		return nil, nil, err
	}

	m := message.Message{
		MessageId: int32(msg.Update().Message.ID),
		UserId:    int32(msg.Update().Message.Sender.ID),
		Type:      "text",
		Content:   []byte(msg.Text()),
		Timestamp: msg.Update().Message.Unixtime,
	}

	bytes, err := proto.Marshal(&m)
	if err != nil {
		return nil, nil, err
	}
	return bytes, nil, nil
}

func validate(msg tele.Context) error {
	if msg == nil {
		return errors.New("message validation error: nil message")
	}

	if msg.Update().Message == nil {
		return errors.New("message validation error: nil update.message")
	}

	if msg.Update().Message.Sender == nil {
		return errors.New("message validation error: nil update.message.sender")
	}

	if strings.TrimSpace(msg.Text()) == "" {
		return errors.New("message validation error: empty message")
	}
	return nil
}

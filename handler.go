package main

import (
	"context"
	"errors"
	"fmt"

	tele "gopkg.in/telebot.v3"
)

type Handler interface {
	Handle(ctx context.Context, msg tele.Context) ([]byte, fmt.Stringer, error)
}

type TextMessageHandler struct{}

var ErrEmptyMessage = errors.New("empty message")

func (h *TextMessageHandler) Handle(ctx context.Context, msg tele.Context) ([]byte, fmt.Stringer, error) {
	txt := msg.Text()
	if txt == "" {
		return nil, nil, ErrEmptyMessage
	}

	return []byte(txt), nil, nil
}

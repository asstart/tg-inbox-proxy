package main

import (
	"context"
	"errors"
	"fmt"

	tele "gopkg.in/telebot.v3"
)

type Handler interface {
	Handle(msg tele.Context, ctx context.Context) ([]byte, fmt.Stringer, error)
}

type TextMessageHandler struct {}

var ErrEmptyMessage = errors.New("empty message")

func (h *TextMessageHandler) Handle(msg tele.Context, ctx context.Context) ([]byte, fmt.Stringer, error) {
	txt := msg.Text()
	if txt == "" {
		return nil, nil, ErrEmptyMessage
	}

	return []byte(txt), nil, nil
}
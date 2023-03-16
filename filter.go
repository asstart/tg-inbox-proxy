package main

import (
	"context"

	tele "gopkg.in/telebot.v3"
)

type Filter interface {
	Filter(msg tele.Context, ctx context.Context) error
}
package main

import (
	"context"

	tele "gopkg.in/telebot.v3"
)

type Pipe struct {
	Filters []Filter
	Handler Handler
	Sender  Sender
}

func (p *Pipe) Process(ctx context.Context, msg tele.Context) error {
	for _, filter := range p.Filters {
		if err := filter.Filter(msg, nil); err != nil {
			return err
		}
	}

	bytes, key, err := p.Handler.Handle(ctx, msg)

	if err != nil {
		return err
	}

	if err := p.Sender.Send(ctx, bytes, key); err != nil {
		return err
	}

	return nil
}

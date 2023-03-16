package main

import (
	"context"
	"log"

	tele "gopkg.in/telebot.v3"
)

type Pipe struct {
	Filters []Filter
	Handler Handler
	Sender  Sender
}

func (p *Pipe) Process(msg tele.Context, ctx context.Context) error {
	log.Print("processing message")
	log.Printf("message: %v", msg.Text())
	for _, filter := range p.Filters {
		if err := filter.Filter(msg, nil); err != nil {
			return err
		}
	}

	log.Print("message passed all filters")

	bytes, key, err := p.Handler.Handle(msg, ctx)

	log.Print("message handled")

	if err != nil {
		return err
	}

	if err = p.Sender.Send(bytes, key, ctx); err != nil {
		return err
	}

	log.Print("message sent")
	
	return nil
}
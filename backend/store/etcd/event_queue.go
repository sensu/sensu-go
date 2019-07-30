package etcd

import (
	"context"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

type EventQueue struct {
	queue chan *corev2.Event
}

func NewEventQueue() *EventQueue {
	return &EventQueue{
		queue: make(chan *corev2.Event, 1000),
	}
}

func (e *EventQueue) Send(ctx context.Context, event *corev2.Event) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case e.queue <- event:
	}
	return nil
}

func (e *EventQueue) Receive(ctx context.Context) (*corev2.Event, error) {
	select {
	case e := <-e.queue:
		return e, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

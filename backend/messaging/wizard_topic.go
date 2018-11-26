package messaging

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/sensu/sensu-go/backend/ring"
	"github.com/sensu/sensu-go/util/retry"

	"github.com/sensu/sensu-go/types"
)

// wizardTopic encapsulates state around a WizardBus topic and its
// consumer channel bindings.
type wizardTopic struct {
	id       string
	bindings map[string]Subscriber
	ring     types.Ring
	sync.RWMutex
	done chan struct{}
}

// Send a message to all subscribers to this topic.
func (t *wizardTopic) Send(msg interface{}) {
	t.RLock()
	subscribers := make([]Subscriber, 0, len(t.bindings))
	for _, subscriber := range t.bindings {
		subscribers = append(subscribers, subscriber)
	}
	t.RUnlock()
	for _, subscriber := range subscribers {
		select {
		case subscriber.Receiver() <- msg:
		case <-t.done:
			return
		}
	}
}

// SoundRoundRobin sends a message to the next subscriber in a round-robin
// ring. In a distributed environment, SendDirect may send a message, or it
// may not, depending if the next subscriber in the round-robin is bound to
// the backend.
func (t *wizardTopic) SendRoundRobin(ctx context.Context, msg interface{}) error {
	if t.ring == nil {
		return errors.New("no ring for topic: " + t.id)
	}

	id, err := t.ring.Next(ctx)
	if err != nil {
		if err == ring.ErrNotOwner || err == ring.ErrEmptyRing {
			return nil
		}
		return err
	}

	t.RLock()
	receiver := t.bindings[id].Receiver()
	t.RUnlock()

	receiver <- msg

	return nil
}

// Subscribe a Subscriber to this topic and receive a Subscription.
func (t *wizardTopic) Subscribe(id string, sub Subscriber) (Subscription, error) {
	t.Lock()
	t.bindings[id] = sub
	t.Unlock()

	if t.ring != nil {
		if err := t.ring.Add(context.Background(), id); err != nil {
			return Subscription{}, err
		}
	}

	return Subscription{
		id:     id,
		cancel: t.unsubscribe,
	}, nil
}

// Unsubscribe a consumer from this topic.
func (t *wizardTopic) unsubscribe(id string) error {
	t.Lock()
	delete(t.bindings, id)
	if len(t.bindings) == 0 {
		select {
		case <-t.done:
		default:
			close(t.done)
		}
	}
	t.Unlock()

	if t.ring != nil {
		backoff := &retry.ExponentialBackoff{
			MaxDelayInterval: 5 * time.Second,
			MaxElapsedTime:   60 * time.Second,
		}

		var err error
		backoff.Retry(func(int) (bool, error) {
			err = t.ring.Remove(context.Background(), id)
			if err != nil && err != ring.ErrNotOwner {
				return false, nil
			}
			return true, nil
		})
		return err
	}

	return nil
}

// Close all WizardTopic bindings.
func (t *wizardTopic) Close() {
	t.Lock()
	select {
	case <-t.done:
		return
	default:
	}
	close(t.done)
	for consumer := range t.bindings {
		delete(t.bindings, consumer)
	}
	t.Unlock()
}

func (t *wizardTopic) IsClosed() bool {
	select {
	case <-t.done:
		return true
	default:
		return false
	}
}

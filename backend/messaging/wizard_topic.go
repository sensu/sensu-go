package messaging

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Sirupsen/logrus"
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
	droppedMessages int64
	done            chan struct{}
}

func (t *wizardTopic) logDroppedMessages() {
	dropped := atomic.LoadInt64(&t.droppedMessages)
	if dropped == 0 {
		return
	}
	// Inexact but avoids contention; good enough for our purposes.
	atomic.StoreInt64(&t.droppedMessages, 0)
	logger.WithFields(logrus.Fields{
		"topic": t.id,
		"dropped_messages_per_second": dropped}).Warn(
		"not all messages could be delivered")
}

// Send a message to all subscribers to this topic.
func (t *wizardTopic) Send(msg interface{}) {
	t.RLock()
	names := make([]string, 0, len(t.bindings))
	subscribers := make([]Subscriber, 0, len(t.bindings))
	for name, subscriber := range t.bindings {
		names = append(names, name)
		subscribers = append(subscribers, subscriber)
	}
	t.RUnlock()
	for i, subscriber := range subscribers {
		select {
		case subscriber.Receiver() <- msg:
		case <-time.After(10 * time.Second):
			logger.WithFields(logrus.Fields{
				"topic":      t.id,
				"subscriber": names[i],
			}).Warn("timed out delivering message to subscriber")
		}
	}
}

// SendDirect sends a message directly to a subscriber of this topic.
func (t *wizardTopic) SendDirect(msg interface{}) error {
	if t.ring == nil {
		return errors.New("no ring for topic: " + t.id)
	}

	id, err := t.ring.Next(context.Background())
	if err != nil {
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
	close(t.done)
	t.Lock()
	for consumer := range t.bindings {
		delete(t.bindings, consumer)
	}
	t.Unlock()
}

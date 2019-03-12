package messaging

import (
	"sync"
)

// wizardTopic encapsulates state around a WizardBus topic and its
// consumer channel bindings.
type wizardTopic struct {
	id       string
	bindings map[string]Subscriber
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

// Subscribe a Subscriber to this topic and receive a Subscription.
func (t *wizardTopic) Subscribe(id string, sub Subscriber) (Subscription, error) {
	t.Lock()
	t.bindings[id] = sub
	t.Unlock()

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

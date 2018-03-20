package messaging

import (
	"context"
	"errors"
	"sync"

	"github.com/sensu/sensu-go/types"
)

// wizardTopic encapsulates state around a WizardBus topic and its
// consumer channel bindings.
type wizardTopic struct {
	id       string
	bindings map[string]Subscriber
	ring     types.Ring
	sync.RWMutex
}

// Send a message to all subscribers to this topic.
func (wTopic *wizardTopic) Send(msg interface{}) {
	wTopic.RLock()
	defer wTopic.RUnlock()

	for _, subscriber := range wTopic.bindings {
		select {
		case subscriber.Receiver() <- msg:
		default:
			continue
		}
	}
}

// SendDirect sends a message directly to a subscriber of this topic.
func (wTopic *wizardTopic) SendDirect(msg interface{}) error {
	if wTopic.ring == nil {
		return errors.New("no ring for topic: " + wTopic.id)
	}

	wTopic.RLock()
	defer wTopic.RUnlock()

	id, err := wTopic.ring.Next(context.Background())
	if err != nil {
		return err
	}

	wTopic.bindings[id].Receiver() <- msg

	return nil
}

// Subscribe a Subscriber to this topic and receive a Subscription.
func (wTopic *wizardTopic) Subscribe(id string, sub Subscriber) (Subscription, error) {
	wTopic.Lock()
	defer wTopic.Unlock()

	wTopic.bindings[id] = sub

	if wTopic.ring != nil {
		if err := wTopic.ring.Add(context.Background(), id); err != nil {
			return Subscription{}, err
		}
	}

	return Subscription{
		id:     id,
		cancel: wTopic.unsubscribe,
	}, nil
}

// Unsubscribe a consumer from this topic.
func (wTopic *wizardTopic) unsubscribe(id string) {
	wTopic.Lock()
	delete(wTopic.bindings, id)
	if wTopic.ring != nil {
		// TODO: What do we do in the case of this failing?
		_ = wTopic.ring.Remove(context.Background(), id)
	}
	wTopic.Unlock()
}

// Close all WizardTopic bindings.
func (wTopic *wizardTopic) Close() {
	wTopic.Lock()
	for consumer := range wTopic.bindings {
		delete(wTopic.bindings, consumer)
	}
	wTopic.Unlock()
}

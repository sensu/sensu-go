package messaging

import "sync"

// WizardTopic encapsulates state around a WizardBus topic and its
// consumer channel bindings.
type WizardTopic struct {
	sync.RWMutex
	Bindings map[string]chan<- interface{}
}

// Send a message to all subscribers to this topic.
func (wTopic *WizardTopic) Send(msg interface{}) {
	wTopic.RLock()

	for _, ch := range wTopic.Bindings {
		select {
		case ch <- msg:
		default:
			continue
		}
	}

	wTopic.RUnlock()
}

// Subscribe a channel, identified by a consumer name, to this topic.
func (wTopic *WizardTopic) Subscribe(consumer string, channel chan<- interface{}) {
	wTopic.Lock()
	wTopic.Bindings[consumer] = channel
	wTopic.Unlock()
}

// Unsubscribe a consumer from this topic.
func (wTopic *WizardTopic) Unsubscribe(consumer string) {
	wTopic.Lock()

	delete(wTopic.Bindings, consumer)

	wTopic.Unlock()
}

// Close all WizardTopic bindings.
func (wTopic *WizardTopic) Close() {
	wTopic.Lock()
	for consumer, _ := range wTopic.Bindings {
		delete(wTopic.Bindings, consumer)
	}
	wTopic.Unlock()
}

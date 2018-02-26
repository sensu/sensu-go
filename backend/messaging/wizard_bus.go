package messaging

import (
	"errors"
	"sync"
	"sync/atomic"
)

// WizardBus is an in-memory message bus.
//
// For every topic, WizardBus creates a new goroutine responsible for fanning
// messages out to each subscriber for a given topic. Any type can be passed
// across a WizardTopic and it is up to the consumers/producers to coordinate
// around a particular topic type. Care should be taken not to send multiple
// message types over a single topic, however, as we do not want to introduce
// a dependency on reflection to determine the type of the received interface{}.
type WizardBus struct {
	running *atomic.Value
	mutex   *sync.RWMutex
	errchan chan error
	topics  map[string]*WizardTopic
}

// Start ...
func (b *WizardBus) Start() error {
	b.errchan = make(chan error, 1)
	b.running = &atomic.Value{}
	b.mutex = &sync.RWMutex{}
	b.topics = map[string]*WizardTopic{}
	b.running.Store(true)

	return nil
}

// Stop ...
func (b *WizardBus) Stop() error {
	b.running.Store(false)
	close(b.errchan)
	b.mutex.Lock()
	for _, wTopic := range b.topics {
		wTopic.Close()
	}
	b.mutex.Unlock()
	return nil
}

// Status ...
func (b *WizardBus) Status() error {
	if !b.running.Load().(bool) {
		return errors.New("bus has shutdown")
	}
	return nil
}

// Err ...
func (b *WizardBus) Err() <-chan error {
	return b.errchan
}

// Create a WizardBus topic (WizardTopic) with consumer channel
// bindings. Every topic has its own mutex, sending data to consumers
// should only be blocked when adding (Subscribe) or removing
// (Unsubscribe) a consumer binding to the topic. This function also
// creates a goroutine that will continue to pull a message from the
// topic's send buffer, lock the topic's mutex (R), write the message
// to each consumer channel bound to the topic, and then unlock the
// topic's mutex.
func (b *WizardBus) createTopic(topic string) *WizardTopic {
	wTopic := &WizardTopic{
		Bindings: make(map[string](chan<- interface{})),
	}

	return wTopic
}

// Subscribe to a WizardBus topic. This function locks the WizardBus
// mutex (RW), fetches the appropriate WizardTopic (or creates it if
// missing), unlocks the WizardBus mutex, locks the WizardTopic's
// mutex (RW), adds the consumer channel to the WizardTopic's
// bindings, and unlocks the WizardTopics mutex.
//
// WARNING:
//
// Messages received over a topic should be considered IMMUTABLE by consumers.
// Modifying received messages will introduce data races. While these _may_ be
// detected by the Golang race detector, this is not always the case and is
// only exacerbated by the fact that we test each package individually.
func (b *WizardBus) Subscribe(topic string, consumer string, channel chan<- interface{}) error {
	if !b.running.Load().(bool) {
		return errors.New("bus no longer running")
	}

	b.mutex.Lock()
	defer b.mutex.Unlock()

	if _, ok := b.topics[topic]; !ok {
		b.topics[topic] = b.createTopic(topic)
	}

	b.topics[topic].Subscribe(consumer, channel)

	return nil
}

// Unsubscribe from a WizardBus topic. This function locks the
// WizardBus mutex (RW), fetches the appropriate WizardTopic (noop if
// missing), unlocks the WizardBus mutex, locks the WizardTopic's
// mutex (RW), fetches the consumer channel from the WizardTopic's
// bindings (noop if missing), and deletes the channel from
// WizardTopic's bindings (it does not close the channel because it
// could be bound to another topic).
func (b *WizardBus) Unsubscribe(topic string, consumer string) error {
	if !b.running.Load().(bool) {
		return errors.New("bus no longer running")
	}

	b.mutex.RLock()
	wTopic, ok := b.topics[topic]
	if !ok {
		return errors.New("topic not found")
	}
	b.mutex.RUnlock()

	wTopic.Unsubscribe(consumer)
	return nil
}

// Publish publishes a message to a topic. If the topic does not
// exist, this is a noop.
func (b *WizardBus) Publish(topic string, msg interface{}) error {
	if !b.running.Load().(bool) {
		return errors.New("bus no longer running")
	}

	b.mutex.RLock()
	defer b.mutex.RUnlock()

	if wTopic, ok := b.topics[topic]; ok {
		wTopic.Send(msg)
	}

	return nil
}

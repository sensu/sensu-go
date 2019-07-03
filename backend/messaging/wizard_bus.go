package messaging

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/prometheus/client_golang/prometheus"
)

// WizardBus is a message bus.
//
// For every topic, WizardBus creates a new goroutine responsible for fanning
// messages out to each subscriber for a given topic. Any type can be passed
// across a WizardTopic and it is up to the consumers/producers to coordinate
// around a particular topic type. Care should be taken not to send multiple
// message types over a single topic, however, as we do not want to introduce
// a dependency on reflection to determine the type of the received interface{}.
type WizardBus struct {
	running  atomic.Value
	topicsMu sync.RWMutex
	topics   map[string]*wizardTopic
	errchan  chan error
}

// WizardBusConfig configures a WizardBus
type WizardBusConfig struct{}

// WizardOption is a functional option.
type WizardOption func(*WizardBus) error

// NewWizardBus creates a new WizardBus.
func NewWizardBus(cfg WizardBusConfig, opts ...WizardOption) (*WizardBus, error) {
	bus := &WizardBus{
		errchan: make(chan error, 1),
		topics:  make(map[string]*wizardTopic),
	}
	for _, opt := range opts {
		if err := opt(bus); err != nil {
			return nil, err
		}
	}
	_ = prometheus.Register(topicCounter)

	return bus, nil
}

// Start ...
func (b *WizardBus) Start() error {
	b.running.Store(true)
	return nil
}

// Stop ...
func (b *WizardBus) Stop() error {
	b.running.Store(false)
	close(b.errchan)
	b.topicsMu.Lock()
	for _, wTopic := range b.topics {
		wTopic.Close()
	}
	b.topicsMu.Unlock()
	return nil
}

// Err ...
func (b *WizardBus) Err() <-chan error {
	return b.errchan
}

// Name returns the daemon name
func (b *WizardBus) Name() string {
	return "message_bus"
}

// Create a WizardBus topic (WizardTopic) with consumer channel
// bindings. Every topic has its own mutex, sending data to consumers
// should only be blocked when adding (Subscribe) or removing
// (Unsubscribe) a consumer binding to the topic. This function also
// creates a goroutine that will continue to pull a message from the
// topic's send buffer, lock the topic's mutex (R), write the message
// to each consumer channel bound to the topic, and then unlock the
// topic's mutex.
func (b *WizardBus) createTopic(topic string) *wizardTopic {
	wTopic := &wizardTopic{
		id:       topic,
		bindings: make(map[string]Subscriber),
		done:     make(chan struct{}),
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
func (b *WizardBus) Subscribe(topic string, consumer string, sub Subscriber) (Subscription, error) {
	if !b.running.Load().(bool) {
		return Subscription{}, errors.New("bus no longer running")
	}

	b.topicsMu.Lock()
	defer b.topicsMu.Unlock()

	t, ok := b.topics[topic]
	if !ok || t.IsClosed() {
		t = b.createTopic(topic)
		b.topics[topic] = t
	}

	subscription, err := t.Subscribe(consumer, sub)
	return subscription, err
}

// Publish publishes a message to a topic. If the topic does not
// exist, this is a noop.
func (b *WizardBus) Publish(topic string, msg interface{}) error {
	if !b.running.Load().(bool) {
		return errors.New("bus no longer running")
	}

	b.topicsMu.RLock()
	wTopic, ok := b.topics[topic]
	b.topicsMu.RUnlock()

	if ok {
		wTopic.Send(msg)
	}

	return nil
}

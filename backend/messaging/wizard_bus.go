package messaging

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"github.com/sensu/sensu-go/types"
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
	running    atomic.Value
	topicsMu   sync.RWMutex
	topics     map[string]*WizardTopic
	errchan    chan error
	ringGetter types.RingGetter
	ringCache  map[string]types.Ring
	ringMu     sync.Mutex
}

// WizardBusConfig configures a WizardBus
type WizardBusConfig struct {
	RingGetter types.RingGetter
}

// WizardOption is a functional option.
type WizardOption func(*WizardBus) error

// NewWizardBus creates a new WizardBus.
func NewWizardBus(cfg WizardBusConfig, opts ...WizardOption) (*WizardBus, error) {
	bus := &WizardBus{
		errchan:    make(chan error, 1),
		topics:     make(map[string]*WizardTopic),
		ringGetter: cfg.RingGetter,
		ringCache:  make(map[string]types.Ring),
	}
	for _, opt := range opts {
		if err := opt(bus); err != nil {
			return nil, err
		}
	}
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

	b.topicsMu.Lock()
	defer b.topicsMu.Unlock()

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

	b.topicsMu.RLock()
	wTopic, ok := b.topics[topic]
	if !ok {
		return errors.New("topic not found")
	}
	b.topicsMu.RUnlock()

	wTopic.Unsubscribe(consumer)
	b.ringMu.Lock()
	ring, ok := b.ringCache[topic]
	b.ringMu.Unlock()
	if ok {
		return ring.Remove(context.Background(), consumer)
	}

	return nil
}

// Publish publishes a message to a topic. If the topic does not
// exist, this is a noop.
func (b *WizardBus) Publish(topic string, msg interface{}) error {
	if !b.running.Load().(bool) {
		return errors.New("bus no longer running")
	}

	b.topicsMu.RLock()
	defer b.topicsMu.RUnlock()

	if wTopic, ok := b.topics[topic]; ok {
		wTopic.Send(msg)
	}

	return nil
}

// PublishDirect publishes a message to a single consumer.
func (b *WizardBus) PublishDirect(topic string, msg interface{}) error {
	if !b.running.Load().(bool) {
		return errors.New("bus no longer running")
	}

	b.topicsMu.RLock()
	wt, ok := b.topics[topic]
	b.topicsMu.RUnlock()

	if !ok {
		return nil
	}

	ring, err := b.getRing(topic, wt)
	if err != nil {
		return err
	}

	subscriber, err := ring.Next(context.Background())
	if err != nil {
		return err
	}

	wt.RLock()
	defer wt.RUnlock()

	wt.Bindings[subscriber] <- msg

	return nil
}

// getRing finds a cached ring, or constructs one for the topic.
// rings are lazily constructed; they are not created until the need for one
// is identified by a call to PublishDirect.
func (b *WizardBus) getRing(topic string, wt *WizardTopic) (types.Ring, error) {
	var err error
	b.ringMu.Lock()
	defer b.ringMu.Unlock()
	ring, ok := b.ringCache[topic]
	if !ok {
		ring = b.ringGetter.GetRing(topic)
		wt.RLock()
		bindings := make([]string, 0, len(wt.Bindings))
		for subscriber := range wt.Bindings {
			bindings = append(bindings, subscriber)
		}
		wt.RUnlock()
		for _, subscriber := range bindings {
			if err := ring.Add(context.Background(), subscriber); err != nil {
				return nil, err
			}
		}
		b.ringCache[topic] = ring
	}
	return ring, err
}

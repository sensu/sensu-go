package messaging

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
)

// NewSubscriber returns an initialized Subscriber.
func newSubscriber(id string) *WizardSubscriber {
	return &WizardSubscriber{
		id:      id,
		count:   0,
		channel: make(chan []byte, 100),
	}
}

// A WizardSubscriber is a struct that makes channels friendlier and allows
// us to fan-in messages from multiple topics to a single channel.
//
// Subscribers are not to be reused.
type WizardSubscriber struct {
	// ID must be unique for each Subscriber.
	id      string
	channel chan []byte
	count   int
	sync.Mutex
}

// WizardBus is an in-memory message bus. It's unsafe and riddled with problems.
type WizardBus struct {
	sendBuffers      map[string]chan []byte
	sendBuffersMutex *sync.Mutex
	// {
	// 	"topic": {
	// 		"channel": channel,
	// 		...
	// 	},
	// 	...
	// }
	subscribers map[string]map[string]*WizardSubscriber
	mutex       *sync.Mutex
	stopping    chan struct{}
	running     *atomic.Value
	wg          *sync.WaitGroup
	errchan     chan error
	subRegistry map[string]*WizardSubscriber
}

// Start ...
func (b *WizardBus) Start() error {
	b.stopping = make(chan struct{}, 1)
	b.errchan = make(chan error, 1)
	b.sendBuffersMutex = &sync.Mutex{}
	b.running = &atomic.Value{}
	b.wg = &sync.WaitGroup{}
	b.sendBuffers = map[string](chan []byte){}
	b.subscribers = map[string]map[string]*WizardSubscriber{}
	b.subRegistry = map[string]*WizardSubscriber{}
	b.mutex = &sync.Mutex{}
	b.running.Store(true)

	return nil
}

// Stop ...
func (b *WizardBus) Stop() error {
	b.running.Store(false)
	close(b.stopping)
	close(b.errchan)
	b.wg.Wait()
	for _, topic := range b.subscribers {
		for _, sub := range topic {
			close(sub.channel)
		}
	}
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

func (b *WizardBus) startFanout(topic string) {
	go func() {
		defer b.wg.Done()
		c := b.sendBuffers[topic]
		for {
			select {
			case <-b.stopping:
				close(b.sendBuffers[topic])
				for remaining := range b.sendBuffers[topic] {
					for _, sub := range b.subscribers[topic] {
						select {
						case sub.channel <- remaining:
						default:
							continue
						}
					}
				}
				return
			case msg := <-c:
				for _, sub := range b.subscribers[topic] {
					select {
					case sub.channel <- msg:
					default:
						continue
					}
				}
			}
		}
	}()
}

// NewSubscriber returns a new unique subscriber ID and a corresponding event
// channel for that subscriber.
func (b *WizardBus) NewSubscriber() (string, <-chan []byte, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", nil, err
	}

	b.mutex.Lock()
	defer b.mutex.Unlock()

	sub := newSubscriber(id.String())
	b.subRegistry[sub.id] = sub
	return sub.id, sub.channel, nil
}

// Subscribe is not particularly safe for use by multiple goroutines...
func (b *WizardBus) Subscribe(topic, subscriberID string) error {
	if !b.running.Load().(bool) {
		return errors.New("bus no longer running")
	}

	sub, ok := b.subRegistry[subscriberID]
	if !ok {
		return errors.New("subscriber not found")
	}

	b.mutex.Lock()
	defer b.mutex.Unlock()

	if _, ok := b.sendBuffers[topic]; !ok {
		b.sendBuffers[topic] = make(chan []byte, 100)
		b.wg.Add(1)
		b.startFanout(topic)
	}

	_, ok = b.subscribers[topic]
	if !ok {
		b.subscribers[topic] = map[string]*WizardSubscriber{}
	}

	b.subscribers[topic][sub.id] = sub

	return nil
}

// Publish publishes a message. If the topic has never had subscribers, this is
// a noop.
func (b *WizardBus) Publish(topic string, msg []byte) error {
	// At least _try_ to avoid sending on channels that are closing. This is still
	// khnown to be racey, but we're accepting that for now. We need to protect
	// against panics in the Backend.
	if !b.running.Load().(bool) {
		return errors.New("bus no longer running")
	}

	// This is purposefully racey with Subscribe, because we want to avoid
	// having Publish() synchronize around a mutex since it's called all the
	// time. Put all of the synchronization in Subscribe, and understand that
	// nearly _all_ messages sent through sensu's message bus are periodic and
	// cyclical. If we miss the _very first message_ sent over this topic, it
	// really doesn't matter, because it'll be sent again soon enough.
	if ch, ok := b.sendBuffers[topic]; ok {
		ch <- msg
	}

	return nil
}

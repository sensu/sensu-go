package messaging

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/Sirupsen/logrus"
)

var (
	sendBufferQueueLength = 10

	logger = logrus.WithFields(logrus.Fields{
		"component": "message_bus",
	})
)

// WizardBus is an in-memory message bus.
type WizardBus struct {
	stopping chan struct{}
	running  *atomic.Value
	wg       *sync.WaitGroup
	mutex    *sync.RWMutex
	errchan  chan error
	topics   map[string]WizardTopic
}

// WizardTopic encapsulates state around a WizardBus topic and its
// consumer channel bindings.
type WizardTopic struct {
	mutex      *sync.RWMutex
	bindings   map[string]chan<- interface{}
	sendBuffer chan interface{}
}

// Start ...
func (b *WizardBus) Start() error {
	b.stopping = make(chan struct{}, 1)
	b.errchan = make(chan error, 1)
	b.running = &atomic.Value{}
	b.wg = &sync.WaitGroup{}
	b.mutex = &sync.RWMutex{}
	b.topics = map[string]WizardTopic{}
	b.running.Store(true)

	return nil
}

// Stop ...
func (b *WizardBus) Stop() error {
	b.running.Store(false)
	close(b.stopping)
	close(b.errchan)
	b.wg.Wait()
	for _, wTopic := range b.topics {
		for _, binding := range wTopic.bindings {
			close(binding)
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
		mutex:      &sync.RWMutex{},
		bindings:   make(map[string](chan<- interface{})),
		sendBuffer: make(chan interface{}, sendBufferQueueLength),
	}

	b.wg.Add(1)

	go func() {
		defer b.wg.Done()
		for {
			select {
			case <-b.stopping:
				logger.Infof("message bus - flushing topic %s", topic)
				close(wTopic.sendBuffer)

				wTopic.mutex.RLock()

				for remaining := range wTopic.sendBuffer {
					for _, ch := range wTopic.bindings {
						select {
						case ch <- remaining:
						default:
							continue
						}
					}
				}

				wTopic.mutex.RUnlock()

				return
			case msg := <-wTopic.sendBuffer:
				wTopic.mutex.RLock()

				for _, ch := range wTopic.bindings {
					select {
					case ch <- msg:
					default:
						continue
					}
				}

				wTopic.mutex.RUnlock()
			}
		}
	}()

	return wTopic
}

// Subscribe to a WizardBus topic. This function locks the WizardBus
// mutex (RW), fetches the appropriate WizardTopic (or creates it if
// missing), unlocks the WizardBus mutex, locks the WizardTopic's
// mutex (RW), adds the consumer channel to the WizardTopic's
// bindings, and unlocks the WizardTopics mutex.
func (b *WizardBus) Subscribe(topic string, consumer string, channel chan<- interface{}) error {
	if !b.running.Load().(bool) {
		return errors.New("bus no longer running")
	}

	b.mutex.Lock()

	if _, ok := b.topics[topic]; !ok {
		b.topics[topic] = *b.createTopic(topic)
	}

	b.mutex.Unlock()

	wTopic := b.topics[topic]

	wTopic.mutex.Lock()

	wTopic.bindings[consumer] = channel

	wTopic.mutex.Unlock()

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

	b.mutex.Lock()

	wTopic, ok := b.topics[topic]

	b.mutex.Unlock()

	if ok {
		wTopic.mutex.Lock()

		delete(wTopic.bindings, consumer)

		wTopic.mutex.Unlock()
	}

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
		select {
		case wTopic.sendBuffer <- msg:
			return nil
		default:
			// TODO(greg): we should definitely have metrics around this.
			return errors.New("topic buffer full")
		}
	}

	return nil
}

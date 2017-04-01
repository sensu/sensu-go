package messaging

import (
	"errors"
	"sync"
	"sync/atomic"
)

// WizardBus is an in-memory message bus. It's unsafe and riddled with problems.
type WizardBus struct {
	stopping chan struct{}
	running  *atomic.Value
	wg       *sync.WaitGroup
	mutex    *sync.Mutex
	errchan  chan error
	topics   map[string]WizardTopic
}

// WizardTopic encapsulates state around a WizardBus topic and its
// consumer channel bindings.
type WizardTopic struct {
	mutex      *sync.Mutex
	bindings   map[string]chan<- []byte
	sendBuffer chan []byte
}

// Start ...
func (b *WizardBus) Start() error {
	b.stopping = make(chan struct{}, 1)
	b.errchan = make(chan error, 1)
	b.running = &atomic.Value{}
	b.wg = &sync.WaitGroup{}
	b.mutex = &sync.Mutex{}
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
	for _, wtopic := range b.topics {
		for _, binding := range wtopic.bindings {
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

// Create a WizardBus topic with consumer channel bindings. Every
// topic has its own mutex, so sending data to consumers should only
// block adding (Subscribe) and removing (Unsubscribe) consumers
// to/from the single topic.
func (b *WizardBus) createTopic(topic string) *WizardTopic {
	wTopic := &WizardTopic{
		mutex:      &sync.Mutex{},
		bindings:   make(map[string](chan<- []byte)),
		sendBuffer: make(chan []byte),
	}

	b.wg.Add(1)

	go func() {
		defer b.wg.Done()
		for {
			select {
			case <-b.stopping:
				close(wTopic.sendBuffer)

				wTopic.mutex.Lock()

				for remaining := range wTopic.sendBuffer {
					for _, ch := range wTopic.bindings {
						select {
						case ch <- remaining:
						default:
							continue
						}
					}
				}

				wTopic.mutex.Unlock()

				return
			case msg := <-wTopic.sendBuffer:
				wTopic.mutex.Lock()

				for _, ch := range wTopic.bindings {
					select {
					case ch <- msg:
					default:
						continue
					}
				}

				wTopic.mutex.Unlock()
			}
		}
	}()

	return wTopic
}

// Subscribe to a WizardBus topic. This function locks the WizardBus
// mutex, fetches the appropriate WizardTopic (or creates it if
// missing), unlocks the WizardBus mutex, locks the WizardTopic's
// mutex, adds the consumer channel to the WizardTopic's bindings, and
// unlocks the WizardTopics mutex.
func (b *WizardBus) Subscribe(topic string, consumer string, channel chan<- []byte) error {
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
// WizardBus mutex, fetches the appropriate WizardTopic (noop if
// missing), unlocks the WizardBus mutex, locks the WizardTopic's
// mutex, fetches the consumer channel from the WizardTopic's bindings
// (noop if missing), and closes the channel and deletes it from
// bindings map.
func (b *WizardBus) Unsubscribe(topic string, consumer string) error {
	if !b.running.Load().(bool) {
		return errors.New("bus no longer running")
	}

	b.mutex.Lock()

	wTopic, ok := b.topics[topic]

	b.mutex.Unlock()

	if ok {
		wTopic.mutex.Lock()

		if channel, ok := wTopic.bindings[consumer]; ok {
			close(channel)
			delete(wTopic.bindings, consumer)

		}

		wTopic.mutex.Unlock()
	}

	return nil
}

// Publish publishes a message. If the topic does not exist, this is a
// noop.
func (b *WizardBus) Publish(topic string, msg []byte) error {
	if !b.running.Load().(bool) {
		return errors.New("bus no longer running")
	}

	// This is purposefully racey with Subscribe, because we want
	// to avoid having Publish() synchronize around a mutex since
	// it's called very often. Put all of the synchronization in
	// Subscribe, and understand that nearly _all_ messages sent
	// through Sensu's message bus are periodic and cyclical. If
	// we miss the _very first message_ sent over this topic, it
	// really doesn't matter, because it'll be sent again soon
	// enough.
	if wTopic, ok := b.topics[topic]; ok {
		wTopic.sendBuffer <- msg
	}

	return nil
}

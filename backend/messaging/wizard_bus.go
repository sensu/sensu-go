package messaging

import (
	"errors"
	"sync"
	"sync/atomic"
)

// WizardBus is an in-memory message bus. It's unsafe and riddled with problems.
type WizardBus struct {
	sendBuffers      map[string]chan []byte
	sendBuffersMutex *sync.Mutex
	subscribers      map[string][]chan []byte
	stopping         chan struct{}
	running          *atomic.Value
	wg               *sync.WaitGroup
	errchan          chan error
}

// Start ...
func (b *WizardBus) Start() error {
	b.stopping = make(chan struct{}, 1)
	b.errchan = make(chan error, 1)
	b.sendBuffersMutex = &sync.Mutex{}
	b.running = &atomic.Value{}
	b.wg = &sync.WaitGroup{}
	b.sendBuffers = map[string](chan []byte){}
	b.subscribers = map[string]([]chan []byte){}
	b.running.Store(true)
	return nil
}

// Stop ...
func (b *WizardBus) Stop() error {
	b.running.Store(false)
	close(b.stopping)
	close(b.errchan)
	b.wg.Wait()
	for _, subs := range b.subscribers {
		for _, ch := range subs {
			close(ch)
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
	done := make(chan struct{})
	go func() {
		c := b.sendBuffers[topic]
		close(done)
		for {
			select {
			case <-b.stopping:
				close(b.sendBuffers[topic])
				for remaining := range b.sendBuffers[topic] {
					for _, ch := range b.subscribers[topic] {
						select {
						case ch <- remaining:
						default:
							continue
						}
					}
				}
				b.wg.Done()
				return
			case msg := <-c:
				for _, ch := range b.subscribers[topic] {
					select {
					case ch <- msg:
					default:
						continue
					}
				}
			}
		}
	}()
	<-done
}

// Subscribe ... We actually ignore channel names here, because right now we
// don't need multiple consumers on a single channel. Just move to NSQ.
func (b *WizardBus) Subscribe(topic, channel string) (<-chan []byte, error) {
	if !b.running.Load().(bool) {
		return nil, errors.New("bus no longer running")
	}

	b.sendBuffersMutex.Lock()
	defer b.sendBuffersMutex.Unlock()
	_, ok := b.sendBuffers[topic]
	if !ok {
		b.sendBuffers[topic] = make(chan []byte, 100)
		b.startFanout(topic)
		b.wg.Add(1)
	}

	subscriberChannel := make(chan []byte, 100)
	b.subscribers[topic] = append(b.subscribers[topic], subscriberChannel)

	return subscriberChannel, nil
}

// Publish ...
func (b *WizardBus) Publish(topic string, msg []byte) error {
	if !b.running.Load().(bool) {
		return errors.New("bus no longer running")
	}

	b.sendBuffersMutex.Lock()
	defer b.sendBuffersMutex.Unlock()
	_, ok := b.sendBuffers[topic]
	if !ok {
		b.sendBuffers[topic] = make(chan []byte)
		b.startFanout(topic)
		b.wg.Add(1)
	}

	b.sendBuffers[topic] <- msg

	return nil
}

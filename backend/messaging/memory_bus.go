package messaging

import (
	"errors"
	"strings"
	"sync"
	"sync/atomic"
)

// MemoryBus is an in-memory message bus. It's unsafe and riddled with problems.
type MemoryBus struct {
	sendBuffers   map[string]chan []byte
	channelsMutex *sync.Mutex
	subscribers   map[string][]chan []byte
	stopping      chan struct{}
	running       *atomic.Value
	wg            *sync.WaitGroup
	errchan       chan error
}

// Start ...
func (b *MemoryBus) Start() error {
	b.stopping = make(chan struct{}, 1)
	b.errchan = make(chan error, 1)
	b.channelsMutex = &sync.Mutex{}
	b.running = &atomic.Value{}
	b.wg = &sync.WaitGroup{}
	b.sendBuffers = map[string](chan []byte){}
	b.subscribers = map[string]([]chan []byte){}
	b.running.Store(true)
	return nil
}

// Stop ...
func (b *MemoryBus) Stop() error {
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
func (b *MemoryBus) Status() error {
	if !b.running.Load().(bool) {
		return errors.New("bus has shutdown")
	}
	return nil
}

// Err ...
func (b *MemoryBus) Err() <-chan error {
	return b.errchan
}

func (b *MemoryBus) startFanout(topicName string) {
	done := make(chan struct{})
	go func() {
		c := b.sendBuffers[topicName]
		close(done)
		for {
			select {
			case <-b.stopping:
				close(b.sendBuffers[topicName])
				for remaining := range b.sendBuffers[topicName] {
					for _, subChan := range b.subscribers[topicName] {
						select {
						case subChan <- remaining:
						default:
							continue
						}
					}
				}
				b.wg.Done()
				return
			case msg := <-c:
				for _, subChan := range b.subscribers[topicName] {
					select {
					case subChan <- msg:
					default:
						continue
					}
				}
			}
		}
	}()
	<-done
}

// Subscribe ...
func (b *MemoryBus) Subscribe(topic, channel string) (<-chan []byte, error) {
	if !b.running.Load().(bool) {
		return nil, errors.New("bus no longer running")
	}

	topicName := strings.Join([]string{topic, channel}, ".")

	b.channelsMutex.Lock()
	defer b.channelsMutex.Unlock()
	_, ok := b.sendBuffers[topicName]
	if !ok {
		b.sendBuffers[topicName] = make(chan []byte, 100)
		b.startFanout(topicName)
		b.wg.Add(1)
	}

	subscriberChannel := make(chan []byte, 100)
	b.subscribers[topicName] = append(b.subscribers[topicName], subscriberChannel)

	return subscriberChannel, nil
}

// Publish ...
func (b *MemoryBus) Publish(topic, channel string, msg []byte) error {
	if !b.running.Load().(bool) {
		return errors.New("bus no longer running")
	}

	topicName := strings.Join([]string{topic, channel}, ".")

	b.channelsMutex.Lock()
	defer b.channelsMutex.Unlock()
	_, ok := b.sendBuffers[topicName]
	if !ok {
		b.sendBuffers[topicName] = make(chan []byte)
		b.startFanout(topicName)
		b.wg.Add(1)
	}

	b.sendBuffers[topicName] <- msg

	return nil
}

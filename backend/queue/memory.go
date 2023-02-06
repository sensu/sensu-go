package queue

import (
	"context"
	"sync"
	"time"
)

// NewMemoryClient returns an in memory Client implementation
// for testing purposes.
func NewMemoryClient() Client {
	return &memory{
		data:         make(map[string][]Item),
		pollDuration: time.Second,
	}
}

// memory is an implementation of the queue Client interface.
// Provided for testing purposes.
type memory struct {
	data map[string][]Item
	ctr  int64
	sync.Mutex
	pollDuration time.Duration
}

// Enqueue ...
func (m *memory) Enqueue(_ context.Context, val Item) error {
	m.Lock()
	defer m.Unlock()
	m.data[val.Queue] = append(m.data[val.Queue], val)
	return nil
}

// memoryItem is an item from MemoryQueue.
type memoryItem struct {
	value Item
	q     *memory
}

// Value ...
func (m *memoryItem) Item() Item {
	return m.value
}

// Ack ...
func (m *memoryItem) Ack(context.Context) error {
	return nil
}

// Nack ...
func (m *memoryItem) Nack(context.Context) error {
	m.q.Lock()
	defer m.q.Unlock()
	m.q.data[m.value.Queue] = append(m.q.data[m.value.Queue], m.value)
	return nil
}

// Reserve ...
func (m *memory) Reserve(ctx context.Context, queue string) (Reservation, error) {
	// cheesy blocking algo
	var val Item
	for {
		m.Lock()
		if len(m.data[queue]) == 0 {
			m.Unlock()
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(m.pollDuration):
				continue
			}
		}
		val = m.data[queue][0]
		m.data[queue] = m.data[queue][1:]
		m.Unlock()
		break
	}
	return &memoryItem{
		q:     m,
		value: val,
	}, nil
}

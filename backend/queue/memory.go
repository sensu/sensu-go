package queue

import (
	"context"
	"strings"
	"sync"
	"time"
)

// NewMemoryGetter creates a new MemoryGetter.
func NewMemoryGetter() *MemoryGetter {
	return &MemoryGetter{data: make(map[string]*Memory)}
}

// MemoryGetter is a QueueGetter.
type MemoryGetter struct {
	data map[string]*Memory
}

// GetQueue gets a Memory queue.
func (m *MemoryGetter) GetQueue(path ...string) Interface {
	key := strings.Join(path, "/")
	q, ok := m.data[key]
	if !ok {
		q = &Memory{}
		m.data[key] = q
	}
	return q
}

// Memory is an implementation of queue in memory, provided
// for testing purposes.
type Memory struct {
	data []string
	sync.Mutex
}

// Enqueue ...
func (m *Memory) Enqueue(_ context.Context, val string) error {
	m.Lock()
	defer m.Unlock()
	m.data = append(m.data, val)
	return nil
}

// MemoryItem is an item from MemoryQueue.
type MemoryItem struct {
	value string
	q     *Memory
}

// Value ...
func (m *MemoryItem) Value() string {
	return m.value
}

// Ack ...
func (m *MemoryItem) Ack(context.Context) error {
	return nil
}

// Nack ...
func (m *MemoryItem) Nack(context.Context) error {
	m.q.Lock()
	defer m.q.Unlock()
	m.q.data = append(m.q.data, m.value)
	return nil
}

// Dequeue ...
func (m *Memory) Dequeue(context.Context) (QueueItem, error) {
	// cheesy blocking algo
	var val string
	for {
		m.Lock()
		if len(m.data) == 0 {
			m.Unlock()
			time.Sleep(time.Second)
			continue
		}
		val = m.data[0]
		m.data = m.data[1:]
		m.Unlock()
		break
	}
	return &MemoryItem{
		q:     m,
		value: val,
	}, nil
}

package limiter

import (
	"sync"
)

const (
	// entityLimit is the limit of the total number of entities in a cluster.
	entityLimit = 1000

	// entityHistorySize is the maximum number of counts the entity history should store.
	entityHistorySize = 12
)

// Limiter specifies the Limiter interface.
type Limiter interface {
	Limit() int
	CountHistory() []int
	AddCount(int)
}

// EntityLimiter contains the entity count history.
type EntityLimiter struct {
	entityCountHistory []int
	mu                 sync.Mutex
}

// NewEntityLimiter instantiates a new EntityLimiter.
func NewEntityLimiter() *EntityLimiter {
	return &EntityLimiter{
		entityCountHistory: []int{},
	}
}

// Limit returns the entity limit.
func (e *EntityLimiter) Limit() int {
	return entityLimit
}

// CountHistory returns the count history.
func (e *EntityLimiter) CountHistory() []int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.entityCountHistory
}

// AddCount appends a new entity count to the entity history.
func (e *EntityLimiter) AddCount(i int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.entityCountHistory = append(e.entityCountHistory, i)
	if len(e.entityCountHistory) > entityHistorySize {
		e.entityCountHistory = e.entityCountHistory[1:]
	}
}

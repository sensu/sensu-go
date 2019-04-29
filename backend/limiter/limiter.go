package limiter

import (
	"context"
	"sync"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/backend/store/etcd"
	"github.com/sensu/sensu-go/backend/tessend"
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
	License() bool
}

// EntityLimiter contains the entity count history.
type EntityLimiter struct {
	entityCountHistory []int
	mu                 sync.Mutex
	client             *clientv3.Client
	ctx                context.Context
}

// NewEntityLimiter instantiates a new EntityLimiter.
func NewEntityLimiter(ctx context.Context, client *clientv3.Client) *EntityLimiter {
	return &EntityLimiter{
		entityCountHistory: []int{},
		client:             client,
		ctx:                ctx,
	}
}

// Limit returns the entity limit.
func (e *EntityLimiter) Limit() int {
	return entityLimit
}

// License returns a bool indicating the presence of a license.
func (e *EntityLimiter) License() bool {
	wrapper := &tessend.Wrapper{}
	return etcd.Get(e.ctx, e.client, tessend.LicenseStorePath, wrapper) == nil
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

package etcd

import (
	"time"

	"github.com/sensu/sensu-go/backend/queue"
)

// NewQueue creates a new etcd backed queue.
func (s *Store) NewQueue(name string, timeout time.Duration) queue.Interface {
	return queue.EtcdGetter{s.client}.GetQueue(name, timeout)
}

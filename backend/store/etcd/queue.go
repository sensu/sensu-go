package etcd

import (
	"github.com/sensu/sensu-go/backend/queue"
)

// NewQueue creates a new etcd backed queue.
func (s *Store) NewQueue(name string) queue.Interface {
	return queue.EtcdGetter{Client: s.client}.NewQueue(name)
}

package etcd

import (
	"github.com/sensu/sensu-go/backend/ring"
)

// GetRing gets a named Ring.
func (s *Store) GetRing(path ...string) ring.Interface {
	return ring.EtcdGetter{Client: s.client}.GetRing(path...)
}

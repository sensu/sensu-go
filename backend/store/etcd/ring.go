package etcd

import (
	"github.com/sensu/sensu-go/backend/ring"
	"github.com/sensu/sensu-go/types"
)

// GetRing gets a named Ring.
func (s *Store) GetRing(path ...string) types.Ring {
	return ring.EtcdGetter{Client: s.client}.GetRing(path...)
}

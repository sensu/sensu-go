package etcd

import (
	"context"

	"github.com/coreos/etcd/clientv3"
)

func (s *Store) Health(ctx context.Context) (*clientv3.GetResponse, error) {
	// call etcd health func and return status
	response, err := s.client.Get(ctx, "health")
	return response, err
}

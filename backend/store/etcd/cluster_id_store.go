package etcd

import (
	"context"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/backend/store"
)

const (
	clusterIDPrefix = "cluster_id"
)

var (
	clusterIDKeyBuilder = store.NewKeyBuilder(clusterIDPrefix)
)

// CreateClusterID creates a sensu cluster id
func (s *Store) CreateClusterID(ctx context.Context, id string) error {
	key := clusterIDKeyBuilder.Build("")
	if _, err := s.client.Put(ctx, key, id); err != nil {
		return &store.ErrInternal{Message: err.Error()}
	}

	return nil
}

// GetClusterID gets the sensu cluster id
func (s *Store) GetClusterID(ctx context.Context) (string, error) {
	key := clusterIDKeyBuilder.Build("")
	resp, err := s.client.Get(ctx, key, clientv3.WithLimit(1))
	if err != nil {
		return "", &store.ErrInternal{Message: err.Error()}
	}
	if len(resp.Kvs) == 0 {
		return "", &store.ErrNotFound{Key: key}
	}

	return string(resp.Kvs[0].Value), err
}

package etcd

import (
	"context"
	"path"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/etcd/kvc"
	"go.etcd.io/etcd/client/v3"
)

const (
	keepalivesPathPrefix = "keepalives"
)

func getKeepalivePath(keepalivesPath string, entity *corev2.Entity) string {
	return path.Join(keepalivesPath, entity.Namespace, entity.Name)
}

// DeleteFailingKeepalive deletes a failing KeepaliveRecord.
func (s *Store) DeleteFailingKeepalive(ctx context.Context, entity *corev2.Entity) error {
	err := Delete(ctx, s.client, getKeepalivePath(s.keepalivesPath, entity))
	if _, ok := err.(*store.ErrNotFound); ok {
		err = nil
	}
	return err
}

// GetFailingKeepalives gets all of the failing KeepaliveRecords.
func (s *Store) GetFailingKeepalives(ctx context.Context) ([]*corev2.KeepaliveRecord, error) {
	var resp *clientv3.GetResponse
	err := kvc.Backoff(ctx).Retry(func(n int) (done bool, err error) {
		resp, err = s.client.Get(ctx, s.keepalivesPath, clientv3.WithPrefix())
		return kvc.RetryRequest(n, err)
	})

	if err != nil {
		return nil, err
	}

	keepalives := []*corev2.KeepaliveRecord{}
	for _, kv := range resp.Kvs {
		keepalive := &corev2.KeepaliveRecord{}
		if err := unmarshal(kv.Value, keepalive); err != nil {
			return nil, &store.ErrNotValid{Err: err}
		}
		keepalives = append(keepalives, keepalive)
	}

	return keepalives, nil
}

// UpdateFailingKeepalive updates a failing KeepaliveRecord.
func (s *Store) UpdateFailingKeepalive(ctx context.Context, entity *corev2.Entity, expiration int64) error {
	kr := corev2.NewKeepaliveRecord(entity, expiration)
	return CreateOrUpdate(ctx, s.client, getKeepalivePath(s.keepalivesPath, entity), entity.Namespace, kr)
}

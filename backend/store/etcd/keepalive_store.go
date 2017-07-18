package etcd

import (
	"context"
	"encoding/json"
	"path"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/types"
)

const (
	keepalivesPathPrefix = "keepalives"
)

func (s *etcdStore) UpdateFailingKeepalive(ctx context.Context, entity *types.Entity, expiration int64) error {
	kr := types.NewKeepaliveRecord(entity, expiration)
	krBytes, err := json.Marshal(kr)
	if err != nil {
		return err
	}
	_, err = s.client.Put(context.Background(), path.Join(s.keepalivesPath, entity.Organization, entity.ID), string(krBytes))
	return err
}

func (s *etcdStore) GetFailingKeepalives(ctx context.Context) ([]*types.KeepaliveRecord, error) {
	resp, err := s.client.Get(ctx, path.Join(s.keepalivesPath), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	keepalives := []*types.KeepaliveRecord{}
	for _, kv := range resp.Kvs {
		keepalive := &types.KeepaliveRecord{}
		if err := json.Unmarshal(kv.Value, keepalive); err != nil {
			// if we have a problem deserializing a keepalive record, delete that record
			// ignoring any errors we have along the way.
			s.client.Delete(ctx, string(kv.Key))
			continue
		}
		keepalives = append(keepalives, keepalive)
	}

	return keepalives, nil
}

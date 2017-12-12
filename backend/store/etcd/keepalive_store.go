package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"path"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/types"
)

const (
	keepalivesPathPrefix = "keepalives"
)

func getKeepalivePath(keepalivesPath string, entity *types.Entity) string {
	return path.Join(keepalivesPath, entity.Organization, entity.Environment, entity.ID)
}

func (s *etcdStore) DeleteFailingKeepalive(ctx context.Context, entity *types.Entity) error {
	_, err := s.client.Delete(ctx, getKeepalivePath(s.keepalivesPath, entity))
	return err
}

func (s *etcdStore) GetFailingKeepalives(ctx context.Context) ([]*types.KeepaliveRecord, error) {
	resp, err := s.client.Get(ctx, s.keepalivesPath, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	keepalives := []*types.KeepaliveRecord{}
	for _, kv := range resp.Kvs {
		keepalive := &types.KeepaliveRecord{}
		if err := json.Unmarshal(kv.Value, keepalive); err != nil {
			// if we have a problem deserializing a keepalive record, delete that record
			// ignoring any errors we have along the way.
			if _, err := s.client.Delete(ctx, string(kv.Key)); err != nil {
				logger.Debug(err)
			}
			continue
		}
		keepalives = append(keepalives, keepalive)
	}

	return keepalives, nil
}

func (s *etcdStore) UpdateFailingKeepalive(ctx context.Context, entity *types.Entity, expiration int64) error {
	kr := types.NewKeepaliveRecord(entity, expiration)
	krBytes, err := json.Marshal(kr)
	if err != nil {
		return err
	}

	cmp := clientv3.Compare(clientv3.Version(getEnvironmentsPath(entity.Organization, entity.Environment)), ">", 0)
	req := clientv3.OpPut(getKeepalivePath(s.keepalivesPath, entity), string(krBytes))
	res, err := s.kvc.Txn(ctx).If(cmp).Then(req).Commit()
	if err != nil {
		return err
	}
	if !res.Succeeded {
		return fmt.Errorf(
			"could not create the keepalive for entity %s in environment %s/%s",
			entity.ID,
			entity.Organization,
			entity.Environment,
		)
	}

	return err
}

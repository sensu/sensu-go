package etcd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

const (
	etcdRoot = "/sensu.io"
)

func getEntityPath(id string) string {
	return fmt.Sprintf("%s/entities/%s", etcdRoot, id)
}

func getChecksPath(name string) string {
	return fmt.Sprintf("%s/checks/%s", etcdRoot, name)
}

// Store is an implementation of the sensu-go/backend/store.Store iface.
type etcdStore struct {
	client *clientv3.Client
	kvc    clientv3.KV
	etcd   *Etcd
}

func (s *etcdStore) UpdateEntity(e *types.Entity) error {
	eStr, err := json.Marshal(e)
	if err != nil {
		return err
	}
	_, err = s.kvc.Put(context.TODO(), getEntityPath(e.ID), string(eStr))
	return err
}

func (s *etcdStore) DeleteEntity(e *types.Entity) error {
	_, err := s.kvc.Delete(context.TODO(), getEntityPath(e.ID))
	return err
}

func (s *etcdStore) GetEntityByID(id string) (*types.Entity, error) {
	resp, err := s.kvc.Get(context.TODO(), getEntityPath(id), clientv3.WithLimit(1))
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) != 1 {
		return nil, nil
	}
	entity := &types.Entity{}
	err = json.Unmarshal(resp.Kvs[0].Value, entity)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (s *etcdStore) GetEntities() ([]*types.Entity, error) {
	resp, err := s.kvc.Get(context.TODO(), getEntityPath(""), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	earr := make([]*types.Entity, len(resp.Kvs))
	for i, kv := range resp.Kvs {
		entity := &types.Entity{}
		err = json.Unmarshal(kv.Value, entity)
		if err != nil {
			return nil, err
		}
		earr[i] = entity
	}

	return earr, nil
}

// Checks
func (s *etcdStore) GetChecks() ([]*types.Check, error) {
	resp, err := s.kvc.Get(context.TODO(), getChecksPath(""), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	checksArray := make([]*types.Check, len(resp.Kvs))
	for i, kv := range resp.Kvs {
		check := &types.Check{}
		err = json.Unmarshal(kv.Value, check)
		if err != nil {
			return nil, err
		}
		checksArray[i] = check
	}

	return checksArray, nil
}

func (s *etcdStore) GetCheckByName(name string) (*types.Check, error) {
	resp, err := s.kvc.Get(context.TODO(), getChecksPath(name))
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	checkBytes := resp.Kvs[0].Value
	check := &types.Check{}
	if err := json.Unmarshal(checkBytes, check); err != nil {
		return nil, err
	}

	return check, nil
}

func (s *etcdStore) DeleteCheckByName(name string) error {
	_, err := s.kvc.Delete(context.TODO(), getChecksPath(name))
	return err
}

func (s *etcdStore) UpdateCheck(check *types.Check) error {
	if err := check.Validate(); err != nil {
		return err
	}

	checkBytes, err := json.Marshal(check)
	if err != nil {
		return err
	}

	_, err = s.kvc.Put(context.TODO(), getChecksPath(check.Name), string(checkBytes))
	if err != nil {
		return err
	}

	return nil
}

// NewStore ...
func (e *Etcd) NewStore() (store.Store, error) {
	c, err := e.NewClient()
	if err != nil {
		return nil, err
	}

	store := &etcdStore{
		etcd:   e,
		client: c,
		kvc:    clientv3.NewKV(c),
	}
	return store, nil
}

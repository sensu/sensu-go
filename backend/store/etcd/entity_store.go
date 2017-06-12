package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"path"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/types"
)

const (
	entityPathPrefix = "entities"
)

func getEntityPath(org, id string) string {
	return path.Join(etcdRoot, entityPathPrefix, org, id)
}

func (s *etcdStore) UpdateEntity(e *types.Entity) error {
	if err := e.Validate(); err != nil {
		return err
	}

	eStr, err := json.Marshal(e)
	if err != nil {
		return err
	}
	_, err = s.kvc.Put(context.TODO(), getEntityPath(e.Organization, e.ID), string(eStr))
	return err
}

func (s *etcdStore) DeleteEntity(e *types.Entity) error {
	if err := e.Validate(); err != nil {
		return err
	}
	_, err := s.kvc.Delete(context.TODO(), getEntityPath(e.Organization, e.ID))
	return err
}

func (s *etcdStore) DeleteEntityByID(org, id string) error {
	if org == "" || id == "" {
		return errors.New("must specify organization and id")
	}
	_, err := s.kvc.Delete(context.TODO(), getEntityPath(org, id))
	return err
}

func (s *etcdStore) GetEntityByID(org, id string) (*types.Entity, error) {
	if org == "" || id == "" {
		return nil, errors.New("must specify organization and id")
	}
	resp, err := s.kvc.Get(context.TODO(), getEntityPath(org, id), clientv3.WithLimit(1))
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

// GetEntities takes an optional org argument, an empty string will return
// all entities.
func (s *etcdStore) GetEntities(org string) ([]*types.Entity, error) {
	resp, err := s.kvc.Get(context.TODO(), getEntityPath(org, ""), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return []*types.Entity{}, nil
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

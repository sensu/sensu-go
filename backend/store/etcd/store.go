package etcd

import (
	"context"
	"encoding/json"
	"errors"
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
		return nil, errors.New("not found")
	}
	entity := &types.Entity{}
	err = json.Unmarshal(resp.Kvs[0].Value, entity)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (s *etcdStore) Healthy() bool {
	mapi := clientv3.NewMaintenance(s.client)
	// TODO(greg): what can we do with the response? are there some operational
	// parameters that are useful?
	//
	// https://godoc.org/github.com/coreos/etcd/etcdserver/etcdserverpb#StatusResponse
	_, err := mapi.Status(context.TODO(), s.etcd.cfg.ClientListenURL)
	if err != nil {
		return false
	}
	return true
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

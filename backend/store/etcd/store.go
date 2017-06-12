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

func getMutatorsPath(name string) string {
	return fmt.Sprintf("%s/mutators/%s", etcdRoot, name)
}

// Store is an implementation of the sensu-go/backend/store.Store iface.
type etcdStore struct {
	client *clientv3.Client
	kvc    clientv3.KV
	etcd   *Etcd
}

// Mutators
func (s *etcdStore) GetMutators() ([]*types.Mutator, error) {
	resp, err := s.kvc.Get(context.TODO(), getMutatorsPath(""), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return []*types.Mutator{}, nil
	}

	mutatorsArray := make([]*types.Mutator, len(resp.Kvs))
	for i, kv := range resp.Kvs {
		mutator := &types.Mutator{}
		err = json.Unmarshal(kv.Value, mutator)
		if err != nil {
			return nil, err
		}
		mutatorsArray[i] = mutator
	}

	return mutatorsArray, nil
}

func (s *etcdStore) GetMutatorByName(name string) (*types.Mutator, error) {
	resp, err := s.kvc.Get(context.TODO(), getMutatorsPath(name))
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	mutatorBytes := resp.Kvs[0].Value
	mutator := &types.Mutator{}
	if err := json.Unmarshal(mutatorBytes, mutator); err != nil {
		return nil, err
	}

	return mutator, nil
}

func (s *etcdStore) DeleteMutatorByName(name string) error {
	_, err := s.kvc.Delete(context.TODO(), getMutatorsPath(name))
	return err
}

func (s *etcdStore) UpdateMutator(mutator *types.Mutator) error {
	if err := mutator.Validate(); err != nil {
		return err
	}

	mutatorBytes, err := json.Marshal(mutator)
	if err != nil {
		return err
	}

	_, err = s.kvc.Put(context.TODO(), getMutatorsPath(mutator.Name), string(mutatorBytes))
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

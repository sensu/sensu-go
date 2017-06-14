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
	mutatorsPathPrefix = "mutators"
)

func getMutatorsPath(org, name string) string {
	return path.Join(etcdRoot, mutatorsPathPrefix, org, name)
}

// Mutators gets the list of mutators for an (optional) organization. If org is
// the empty string, GetMutators returns all mutators for all orgs.
func (s *etcdStore) GetMutators(org string) ([]*types.Mutator, error) {
	resp, err := s.kvc.Get(context.TODO(), getMutatorsPath(org, ""), clientv3.WithPrefix())
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

func (s *etcdStore) GetMutatorByName(org, name string) (*types.Mutator, error) {
	if org == "" || name == "" {
		return nil, errors.New("must specify organization and name of mutator")
	}
	resp, err := s.kvc.Get(context.TODO(), getMutatorsPath(org, name))
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

func (s *etcdStore) DeleteMutatorByName(org, name string) error {
	if org == "" || name == "" {
		return errors.New("must specify organization and name of mutator")
	}
	_, err := s.kvc.Delete(context.TODO(), getMutatorsPath(org, name))
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

	_, err = s.kvc.Put(context.TODO(), getMutatorsPath(mutator.Organization, mutator.Name), string(mutatorBytes))
	if err != nil {
		return err
	}

	return nil
}

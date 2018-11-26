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

var (
	mutatorsPathPrefix = "mutators"
	mutatorKeyBuilder  = store.NewKeyBuilder(mutatorsPathPrefix)
)

func getMutatorPath(mutator *types.Mutator) string {
	return mutatorKeyBuilder.WithResource(mutator).Build(mutator.Name)
}

func getMutatorsPath(ctx context.Context, name string) string {
	return mutatorKeyBuilder.WithContext(ctx).Build(name)
}

// DeleteMutatorByName deletes a Mutator by name.
func (s *Store) DeleteMutatorByName(ctx context.Context, name string) error {
	if name == "" {
		return errors.New("must specify name of mutator")
	}

	_, err := s.client.Delete(ctx, getMutatorsPath(ctx, name))
	return err
}

// GetMutators gets the list of mutators for an (optional) namespace. If org is
// the empty string, GetMutators returns all mutators for all orgs.
func (s *Store) GetMutators(ctx context.Context) ([]*types.Mutator, error) {
	resp, err := s.client.Get(ctx, getMutatorsPath(ctx, ""), clientv3.WithPrefix())
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

// GetMutatorByName gets a Mutator by name.
func (s *Store) GetMutatorByName(ctx context.Context, name string) (*types.Mutator, error) {
	if name == "" {
		return nil, errors.New("must specify name of mutator")
	}

	resp, err := s.client.Get(ctx, getMutatorsPath(ctx, name))
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

// UpdateMutator updates a Mutator.
func (s *Store) UpdateMutator(ctx context.Context, mutator *types.Mutator) error {
	if err := mutator.Validate(); err != nil {
		return err
	}

	mutatorBytes, err := json.Marshal(mutator)
	if err != nil {
		return err
	}

	cmp := clientv3.Compare(clientv3.Version(getNamespacePath(mutator.Namespace)), ">", 0)
	req := clientv3.OpPut(getMutatorPath(mutator), string(mutatorBytes))
	res, err := s.client.Txn(ctx).If(cmp).Then(req).Commit()
	if err != nil {
		return err
	}
	if !res.Succeeded {
		return fmt.Errorf(
			"could not create the mutator %s in namespace %s",
			mutator.Name,
			mutator.Namespace,
		)
	}

	return nil
}

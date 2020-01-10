package etcd

import (
	"context"
	"errors"

	"github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"
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

// GetMutatorsPath gets the path of the mutator store.
func GetMutatorsPath(ctx context.Context, name string) string {
	return mutatorKeyBuilder.WithContext(ctx).Build(name)
}

// DeleteMutatorByName deletes a Mutator by name.
func (s *Store) DeleteMutatorByName(ctx context.Context, name string) error {
	if name == "" {
		return &store.ErrNotValid{Err: errors.New("must specify name of mutator")}
	}

	if _, err := s.client.Delete(ctx, GetMutatorsPath(ctx, name)); err != nil {
		return &store.ErrInternal{Message: err.Error()}
	}
	return nil
}

// GetMutators gets the list of mutators for a namespace.
func (s *Store) GetMutators(ctx context.Context, pred *store.SelectionPredicate) ([]*types.Mutator, error) {
	mutators := []*types.Mutator{}
	err := List(ctx, s.client, GetMutatorsPath, &mutators, pred)
	return mutators, err
}

// GetMutatorByName gets a Mutator by name.
func (s *Store) GetMutatorByName(ctx context.Context, name string) (*types.Mutator, error) {
	if name == "" {
		return nil, &store.ErrNotValid{Err: errors.New("must specify name of mutator")}
	}

	resp, err := s.client.Get(ctx, GetMutatorsPath(ctx, name))
	if err != nil {
		return nil, &store.ErrInternal{Message: err.Error()}
	}
	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	mutatorBytes := resp.Kvs[0].Value
	mutator := &types.Mutator{}
	if err := unmarshal(mutatorBytes, mutator); err != nil {
		return nil, &store.ErrDecode{Err: err}
	}

	return mutator, nil
}

// UpdateMutator updates a Mutator.
func (s *Store) UpdateMutator(ctx context.Context, mutator *types.Mutator) error {
	if err := mutator.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	mutatorBytes, err := proto.Marshal(mutator)
	if err != nil {
		return &store.ErrEncode{Err: err}
	}

	cmp := clientv3.Compare(clientv3.Version(getNamespacePath(mutator.Namespace)), ">", 0)
	req := clientv3.OpPut(getMutatorPath(mutator), string(mutatorBytes))
	res, err := s.client.Txn(ctx).If(cmp).Then(req).Commit()
	if err != nil {
		return &store.ErrInternal{Message: err.Error()}
	}
	if !res.Succeeded {
		return &store.ErrNamespaceMissing{Namespace: mutator.Namespace}
	}

	return nil
}

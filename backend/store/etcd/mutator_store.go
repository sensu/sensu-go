package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
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

// GetMutators gets the list of mutators for a namespace.
func (s *Store) GetMutators(ctx context.Context, pageSize int64, continueToken string) (mutators []*corev2.Mutator, newContinueToken string, err error) {
	opts := []clientv3.OpOption{
		clientv3.WithLimit(pageSize),
	}

	keyPrefix := getMutatorsPath(ctx, "")
	rangeEnd := clientv3.GetPrefixRangeEnd(keyPrefix)
	opts = append(opts, clientv3.WithRange(rangeEnd))

	resp, err := s.client.Get(ctx, path.Join(keyPrefix, continueToken), opts...)
	if err != nil {
		return nil, "", err
	}
	if len(resp.Kvs) == 0 {
		return []*corev2.Mutator{}, "", nil
	}

	for _, kv := range resp.Kvs {
		mutator := &corev2.Mutator{}
		err = json.Unmarshal(kv.Value, mutator)
		if err != nil {
			return nil, "", err
		}

		mutators = append(mutators, mutator)
	}

	if pageSize != 0 && resp.Count > pageSize {
		lastMutator := mutators[len(mutators)-1]
		newContinueToken = computeContinueToken(ctx, lastMutator)
	}

	return mutators, newContinueToken, nil
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

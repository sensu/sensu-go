package etcd

import (
	"context"
	"errors"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
)

var (
	mutatorsPathPrefix = "mutators"
	mutatorKeyBuilder  = store.NewKeyBuilder(mutatorsPathPrefix)
)

func getMutatorPath(mutator *corev2.Mutator) string {
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
func (s *Store) GetMutators(ctx context.Context, pred *store.SelectionPredicate) ([]*corev2.Mutator, error) {
	mutators := []*corev2.Mutator{}
	err := List(ctx, s.client, GetMutatorsPath, &mutators, pred)
	return mutators, err
}

// GetMutatorByName gets a Mutator by name.
func (s *Store) GetMutatorByName(ctx context.Context, name string) (*corev2.Mutator, error) {
	if name == "" {
		return nil, &store.ErrNotValid{Err: errors.New("must specify name of mutator")}
	}

	var mutator corev2.Mutator
	err := Get(ctx, s.client, GetMutatorsPath(ctx, name), &mutator)
	if err != nil {
		if _, ok := err.(*store.ErrNotFound); ok {
			err = nil
		}
		return nil, err
	}

	return &mutator, nil
}

// UpdateMutator updates a Mutator.
func (s *Store) UpdateMutator(ctx context.Context, mutator *corev2.Mutator) error {
	if err := mutator.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	return CreateOrUpdate(ctx, s.client, getMutatorPath(mutator), mutator.Namespace, mutator)
}

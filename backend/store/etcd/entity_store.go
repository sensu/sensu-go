package etcd

import (
	"context"
	"errors"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
)

const (
	entityPathPrefix = "entities"
)

var (
	entityKeyBuilder = store.NewKeyBuilder(entityPathPrefix)
)

func getEntityPath(entity *corev2.Entity) string {
	return entityKeyBuilder.WithResource(entity).Build(entity.Name)
}

// GetEntitiesPath gets the path of the entity store
func GetEntitiesPath(ctx context.Context, name string) string {
	return entityKeyBuilder.WithContext(ctx).Build(name)
}

// DeleteEntity deletes an Entity.
func (s *Store) DeleteEntity(ctx context.Context, e *corev2.Entity) error {
	if err := e.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}
	err := Delete(ctx, s.client, getEntityPath(e))
	if err != nil {
		if _, ok := err.(*store.ErrNotFound); ok {
			err = nil
		}
	}
	return err
}

// DeleteEntityByName deletes an Entity by its name.
func (s *Store) DeleteEntityByName(ctx context.Context, name string) error {
	if name == "" {
		return &store.ErrNotValid{Err: errors.New("must specify name")}
	}

	key := GetEntitiesPath(ctx, name)
	err := Delete(ctx, s.client, key)
	if err != nil {
		if _, ok := err.(*store.ErrNotFound); ok {
			err = nil
		}
	}
	return err
}

// GetEntityByName gets an Entity by its name.
func (s *Store) GetEntityByName(ctx context.Context, name string) (*corev2.Entity, error) {
	if name == "" {
		return nil, &store.ErrNotValid{Err: errors.New("must specify name")}
	}

	var entity corev2.Entity
	if err := Get(ctx, s.client, GetEntitiesPath(ctx, name), &entity); err != nil {
		if _, ok := err.(*store.ErrNotFound); ok {
			return nil, nil
		}
		return nil, err
	}
	if entity.Labels == nil {
		entity.Labels = make(map[string]string)
	}
	if entity.Annotations == nil {
		entity.Annotations = make(map[string]string)
	}
	return &entity, nil
}

// GetEntities returns the entities for the namespace in the supplied context.
func (s *Store) GetEntities(ctx context.Context, pred *store.SelectionPredicate) ([]*corev2.Entity, error) {
	entities := []*corev2.Entity{}
	err := List(ctx, s.client, GetEntitiesPath, &entities, pred)
	return entities, err
}

// UpdateEntity updates an Entity.
func (s *Store) UpdateEntity(ctx context.Context, e *corev2.Entity) error {
	if err := e.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	return CreateOrUpdate(ctx, s.client, getEntityPath(e), e.Namespace, e)
}

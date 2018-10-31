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
	entityPathPrefix = "entities"
)

var (
	entityKeyBuilder = store.NewKeyBuilder(entityPathPrefix)
)

func getEntityPath(entity *types.Entity) string {
	return entityKeyBuilder.WithResource(entity).Build(entity.ID)
}

func getEntitiesPath(ctx context.Context, id string) string {
	return entityKeyBuilder.WithContext(ctx).Build(id)
}

// DeleteEntity deletes an Entity.
func (s *Store) DeleteEntity(ctx context.Context, e *types.Entity) error {
	if err := e.Validate(); err != nil {
		return err
	}
	_, err := s.client.Delete(ctx, getEntityPath(e))
	return err
}

// DeleteEntityByID deletes an Entity by its ID.
func (s *Store) DeleteEntityByID(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("must specify id")
	}

	_, err := s.client.Delete(ctx, getEntitiesPath(ctx, id))
	return err
}

// GetEntityByID gets an Entity by ID.
func (s *Store) GetEntityByID(ctx context.Context, id string) (*types.Entity, error) {
	if id == "" {
		return nil, errors.New("must specify id")
	}

	resp, err := s.client.Get(ctx, getEntitiesPath(ctx, id), clientv3.WithLimit(1))
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
func (s *Store) GetEntities(ctx context.Context) ([]*types.Entity, error) {
	resp, err := query(ctx, s, getEntitiesPath)
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

// UpdateEntity updates an Entity.
func (s *Store) UpdateEntity(ctx context.Context, e *types.Entity) error {
	if err := e.Validate(); err != nil {
		return err
	}

	eStr, err := json.Marshal(e)
	if err != nil {
		return err
	}

	cmp := clientv3.Compare(clientv3.Version(getNamespacePath(e.Namespace)), ">", 0)
	req := clientv3.OpPut(getEntityPath(e), string(eStr))
	res, err := s.client.Txn(ctx).If(cmp).Then(req).Commit()
	if err != nil {
		return err
	}
	if !res.Succeeded {
		return fmt.Errorf(
			"could not create the entity %s in namespace %s",
			e.ID,
			e.Namespace,
		)
	}

	return nil
}

package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
)

const (
	entityPathPrefix = "entities"
)

var (
	entityKeyBuilder = store.NewKeyBuilder(entityPathPrefix)
)

func getEntityPath(entity *v2.Entity) string {
	return entityKeyBuilder.WithResource(entity).Build(entity.Name)
}

func getEntitiesPath(ctx context.Context, name string) string {
	return entityKeyBuilder.WithContext(ctx).Build(name)
}

// DeleteEntity deletes an Entity.
func (s *Store) DeleteEntity(ctx context.Context, e *v2.Entity) error {
	if err := e.Validate(); err != nil {
		return err
	}
	_, err := s.client.Delete(ctx, getEntityPath(e))
	return err
}

// DeleteEntityByName deletes an Entity by its name.
func (s *Store) DeleteEntityByName(ctx context.Context, name string) error {
	if name == "" {
		return errors.New("must specify name")
	}

	_, err := s.client.Delete(ctx, getEntitiesPath(ctx, name))
	return err
}

// GetEntityByName gets an Entity by its name.
func (s *Store) GetEntityByName(ctx context.Context, name string) (*v2.Entity, error) {
	if name == "" {
		return nil, errors.New("must specify name")
	}

	resp, err := s.client.Get(ctx, getEntitiesPath(ctx, name), clientv3.WithLimit(1))
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) != 1 {
		return nil, nil
	}
	entity := &v2.Entity{}
	err = json.Unmarshal(resp.Kvs[0].Value, entity)
	if err != nil {
		return nil, err
	}
	if entity.Labels == nil {
		entity.Labels = make(map[string]string)
	}
	if entity.Annotations == nil {
		entity.Annotations = make(map[string]string)
	}
	return entity, nil
}

// GetEntities takes an optional org argument, an empty string will return
// all entities.
func (s *Store) GetEntities(ctx context.Context) ([]*v2.Entity, error) {
	resp, err := query(ctx, s, getEntitiesPath)
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return []*v2.Entity{}, nil
	}

	earr := make([]*v2.Entity, len(resp.Kvs))
	for i, kv := range resp.Kvs {
		entity := &v2.Entity{}
		err = json.Unmarshal(kv.Value, entity)
		if err != nil {
			return nil, err
		}
		if entity.Labels == nil {
			entity.Labels = make(map[string]string)
		}
		if entity.Annotations == nil {
			entity.Annotations = make(map[string]string)
		}
		earr[i] = entity
	}

	return earr, nil
}

// UpdateEntity updates an Entity.
func (s *Store) UpdateEntity(ctx context.Context, e *v2.Entity) error {
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
			e.Name,
			e.Namespace,
		)
	}

	return nil
}

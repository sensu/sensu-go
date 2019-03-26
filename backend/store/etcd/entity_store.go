package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/backend/store"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
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

func getEntitiesPath(ctx context.Context, name string) string {
	return entityKeyBuilder.WithContext(ctx).Build(name)
}

// DeleteEntity deletes an Entity.
func (s *Store) DeleteEntity(ctx context.Context, e *corev2.Entity) error {
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
func (s *Store) GetEntityByName(ctx context.Context, name string) (*corev2.Entity, error) {
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
	entity := &corev2.Entity{}
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

// GetEntities returns the entities for the namespace in the supplied context.
func (s *Store) GetEntities(ctx context.Context, pageSize int64, continueToken string) (entities []*corev2.Entity, nextContinueToken string, err error) {
	opts := []clientv3.OpOption{
		clientv3.WithLimit(pageSize),
	}

	keyPrefix := getEntitiesPath(ctx, "")
	rangeEnd := clientv3.GetPrefixRangeEnd(keyPrefix)
	opts = append(opts, clientv3.WithRange(rangeEnd))

	resp, err := s.client.Get(ctx, path.Join(keyPrefix, continueToken), opts...)
	if err != nil {
		return nil, "", err
	}
	if len(resp.Kvs) == 0 {
		return []*corev2.Entity{}, "", nil
	}

	for _, kv := range resp.Kvs {
		entity := &corev2.Entity{}
		err = json.Unmarshal(kv.Value, entity)
		if err != nil {
			return nil, "", err
		}
		if entity.Labels == nil {
			entity.Labels = make(map[string]string)
		}
		if entity.Annotations == nil {
			entity.Annotations = make(map[string]string)
		}

		entities = append(entities, entity)
	}

	if pageSize != 0 && resp.Count > pageSize {
		lastEntity := entities[len(entities)-1]
		nextContinueToken = computeContinueToken(ctx, lastEntity)
	}

	return entities, nextContinueToken, nil
}

// UpdateEntity updates an Entity.
func (s *Store) UpdateEntity(ctx context.Context, e *corev2.Entity) error {
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

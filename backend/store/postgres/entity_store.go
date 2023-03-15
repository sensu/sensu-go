package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/poll"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

type EntityStore struct {
	db DBI
}

func NewEntityStore(db DBI) *EntityStore {
	return &EntityStore{
		db: db,
	}
}

func prepareEntityStores(ctx context.Context, db DBI) (*EntityConfigStore, *EntityStateStore, func(*bool), error) {
	tx, err := db.Begin(ctx)
	if err != nil {
		return nil, nil, nil, err
	}

	cleanup := func(rollback *bool) {
		if rollback == nil {
			tx.Rollback(ctx)
			return
		}
		if *rollback {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
	}

	entityConfigStore := NewEntityConfigStore(tx)
	entityStateStore := NewEntityStateStore(tx)

	return entityConfigStore, entityStateStore, cleanup, nil
}

// DeleteEntity deletes an entity using the given entity struct.
func (s *EntityStore) DeleteEntity(ctx context.Context, entity *corev2.Entity) error {
	if err := entity.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	namespace := entity.GetNamespace()
	name := entity.GetName()
	return s.deleteEntity(ctx, name, namespace)
}

// DeleteEntityByName deletes an entity using the given name and the
// namespace stored in ctx.
func (s *EntityStore) DeleteEntityByName(ctx context.Context, name string) error {
	if name == "" {
		return &store.ErrNotValid{Err: errors.New("must specify name")}
	}

	return s.deleteEntity(ctx, name, corev2.ContextNamespace(ctx))
}

// DeleteEntityByName deletes an entity using the given name and the
// namespace stored in ctx.
func (s *EntityStore) deleteEntity(ctx context.Context, name, namespace string) error {
	entityConfigStore, entityStateStore, cleanup, err := prepareEntityStores(ctx, s.db)
	if err != nil {
		return err
	}
	var rollback bool
	defer cleanup(&rollback)

	if err := entityConfigStore.Delete(ctx, namespace, name); err != nil {
		var e *store.ErrNotFound
		if !errors.As(err, &e) {
			rollback = true
			return err
		}
	}

	if err := entityStateStore.Delete(ctx, namespace, name); err != nil {
		var e *store.ErrNotFound
		if !errors.As(err, &e) {
			rollback = true
			return err
		}
	}

	return nil
}

type uniqueResource struct {
	Name      string
	Namespace string
}

// GetEntities returns all entities in the given ctx's namespace. A nil slice
// with no error is returned if none were found.
func (s *EntityStore) GetEntities(ctx context.Context, pred *store.SelectionPredicate) ([]*corev2.Entity, error) {
	namespace := corev2.ContextNamespace(ctx)

	entityConfigStore, entityStateStore, cleanup, err := prepareEntityStores(ctx, s.db)
	if err != nil {
		return nil, err
	}
	var rollback bool
	defer cleanup(&rollback)

	// Fetch the entity configs with the selection predicate
	configs, err := entityConfigStore.List(ctx, namespace, pred)
	if err != nil {
		rollback = true
		return nil, err
	}

	// Build a list of entity states to fetch from the list of entity configs
	resources := namespacedResourceNames{}
	for _, config := range configs {
		resources[namespace] = append(resources[namespace], config.Metadata.Name)
	}

	// Fetch the entity states using the list of namespaced entity names
	states, err := entityStateStore.GetMultiple(ctx, resources)
	if err != nil {
		rollback = true
		return nil, err
	}

	return entitiesFromConfigsAndStates(configs, states)
}

// Create a list of corev2.Entity values from corev3 configs & states
func entitiesFromConfigsAndStates(configs []*corev3.EntityConfig, states uniqueEntityStates) ([]*corev2.Entity, error) {
	entities := []*corev2.Entity{}
	for _, config := range configs {
		res := uniqueResource{
			Name:      config.Metadata.Name,
			Namespace: config.Metadata.Namespace,
		}
		if state, ok := states[res]; ok {
			entity, err := corev3.V3EntityToV2(config, state)
			if err != nil {
				return nil, &store.ErrNotValid{Err: err}
			}
			entities = append(entities, entity)
		} else {
			// there is a config without a corresponding state, create anyways
			entities = append(entities, entityFromConfigOnly(config))
		}
	}
	return entities, nil
}

func entityFromConfigOnly(config *corev3.EntityConfig) *corev2.Entity {
	state := corev3.NewEntityState(config.Metadata.Namespace, config.Metadata.Name)
	entity, _ := corev3.V3EntityToV2(config, state)
	return entity
}

// GetEntityByName returns an entity using the given name and the namespace stored
// in ctx. The resulting entity is nil if none was found.
func (s *EntityStore) GetEntityByName(ctx context.Context, name string) (*corev2.Entity, error) {
	if name == "" {
		return nil, &store.ErrNotValid{Err: errors.New("must specify name")}
	}

	namespace := corev2.ContextNamespace(ctx)

	entityConfigStore, entityStateStore, cleanup, err := prepareEntityStores(ctx, s.db)
	if err != nil {
		return nil, err
	}
	var rollback bool
	defer cleanup(&rollback)

	cfg, err := entityConfigStore.Get(ctx, namespace, name)
	if err != nil {
		var errNotFound *store.ErrNotFound
		if errors.As(err, &errNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("error fetching entity config: %w", err)
	}
	if cfg == nil {
		return nil, nil
	}

	state, err := entityStateStore.Get(ctx, namespace, name)
	if err != nil {
		var errNotFound *store.ErrNotFound
		if errors.As(err, &errNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("error fetching entity state: %w", err)
	}
	if state == nil {
		return entityFromConfigOnly(cfg), nil
	}

	return corev3.V3EntityToV2(cfg, state)
}

// UpdateEntity creates or updates a given entity.
func (s *EntityStore) UpdateEntity(ctx context.Context, entity *corev2.Entity) error {
	namespace := entity.Namespace
	if namespace == "" {
		entity.Namespace = corev2.ContextNamespace(ctx)
	}

	cfg, state := corev3.V2EntityToV3(entity)

	entityConfigStore, entityStateStore, cleanup, err := prepareEntityStores(ctx, s.db)
	if err != nil {
		return err
	}
	var rollback bool
	defer cleanup(&rollback)

	if err := entityConfigStore.CreateOrUpdate(ctx, cfg); err != nil {
		return fmt.Errorf("error updating entity config: %w", err)
	}
	if err := entityStateStore.CreateOrUpdate(ctx, state); err != nil {
		return fmt.Errorf("error updating entity state: %w", err)
	}
	return nil
}

type EntityConfigPoller struct {
	db  DBI
	req storev2.ResourceRequest
}

func (e *EntityConfigPoller) Now(ctx context.Context) (time.Time, error) {
	var now time.Time
	row := e.db.QueryRow(ctx, "SELECT NOW();")
	if err := row.Scan(&now); err != nil {
		return now, &store.ErrInternal{Message: err.Error()}
	}
	return now, nil
}

func (e *EntityConfigPoller) Since(ctx context.Context, updatedSince time.Time) ([]poll.Row, error) {
	wrapper := &EntityConfigWrapper{
		Namespace: e.req.Namespace,
		Name:      e.req.Name,
		UpdatedAt: updatedSince,
	}
	queryParams := wrapper.SQLParams()
	rows, rerr := e.db.Query(ctx, pollEntityConfigQuery, queryParams...)
	if rerr != nil {
		logger.Errorf("entity config since query failed with error %v", rerr)
		return nil, &store.ErrInternal{Message: rerr.Error()}
	}
	defer rows.Close()
	var since []poll.Row
	for rows.Next() {
		if err := rows.Scan(wrapper.SQLParams()...); err != nil {
			return nil, &store.ErrInternal{Message: err.Error()}
		}
		if err := rows.Err(); err != nil {
			return nil, &store.ErrInternal{Message: err.Error()}
		}
		id := fmt.Sprintf("%s/%s", wrapper.Namespace, wrapper.Name)
		pollResult := poll.Row{
			Id:        id,
			Resource:  wrapper,
			CreatedAt: wrapper.CreatedAt,
			UpdatedAt: wrapper.UpdatedAt,
		}
		if wrapper.DeletedAt.Valid {
			pollResult.DeletedAt = &wrapper.DeletedAt.Time
		}
		since = append(since, pollResult)
	}
	return since, nil
}

func newEntityConfigPoller(req storev2.ResourceRequest, db DBI) (poll.Table, error) {
	return &EntityConfigPoller{db: db, req: req}, nil
}

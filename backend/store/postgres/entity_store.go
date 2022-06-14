package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/backend/store/v2/etcdstore"
	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type EntityStore struct {
	store *StoreV2
}

func NewEntityStore(db *pgxpool.Pool, client *clientv3.Client) *EntityStore {
	return &EntityStore{
		store: NewStoreV2(db, client),
	}
}

// DeleteEntity deletes an entity using the given entity struct.
func (e *EntityStore) DeleteEntity(ctx context.Context, entity *corev2.Entity) error {
	if err := entity.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}
	state := &corev3.EntityState{
		Metadata: &entity.ObjectMeta,
	}
	config := &corev3.EntityConfig{
		Metadata: &entity.ObjectMeta,
	}
	stateReq := storev2.NewResourceRequestFromResource(ctx, state)
	stateReq.UsePostgres = true
	configReq := storev2.NewResourceRequestFromResource(ctx, config)
	configReq.UsePostgres = true

	if err := e.store.Delete(configReq); err != nil {
		if _, ok := err.(*store.ErrNotFound); !ok {
			return err
		}
	}
	if err := e.store.Delete(stateReq); err != nil {
		if _, ok := err.(*store.ErrNotFound); ok {
			return nil
		}
		return err
	}
	return nil
}

// DeleteEntityByName deletes an entity using the given name and the
// namespace stored in ctx.
func (e *EntityStore) DeleteEntityByName(ctx context.Context, name string) error {
	if name == "" {
		return &store.ErrNotValid{Err: errors.New("must specify name")}
	}

	state := &corev3.EntityState{
		Metadata: &corev2.ObjectMeta{
			Name:      name,
			Namespace: corev2.ContextNamespace(ctx),
		},
	}
	config := &corev3.EntityConfig{
		Metadata: &corev2.ObjectMeta{
			Name:      name,
			Namespace: corev2.ContextNamespace(ctx),
		},
	}
	stateReq := storev2.NewResourceRequestFromResource(ctx, state)
	stateReq.UsePostgres = true
	configReq := storev2.NewResourceRequestFromResource(ctx, config)
	configReq.UsePostgres = true
	if err := e.store.Delete(configReq); err != nil {
		if _, ok := err.(*store.ErrNotFound); !ok {
			return err
		}
	}
	if err := e.store.Delete(stateReq); err != nil {
		if _, ok := err.(*store.ErrNotFound); ok {
			return nil
		}
		return err
	}
	return nil
}

type uniqueResource struct {
	Name      string
	Namespace string
}

// GetEntities returns all entities in the given ctx's namespace. A nil slice
// with no error is returned if none were found.
func (e *EntityStore) GetEntities(ctx context.Context, pred *store.SelectionPredicate) ([]*corev2.Entity, error) {
	namespace := corev2.ContextNamespace(ctx)

	// Fetch the entity configs with the selection predicate
	configReq := storev2.ResourceRequest{
		Namespace:   namespace,
		Context:     ctx,
		StoreName:   new(corev3.EntityConfig).StoreName(),
		UsePostgres: true,
	}
	if pred.Ordering == corev2.EntitySortName {
		configReq.SortOrder = storev2.SortAscend
		if pred.Descending {
			configReq.SortOrder = storev2.SortDescend
		}
	}

	wConfigs, err := e.store.List(configReq, pred)
	if err != nil {
		return nil, err
	}
	configs := make([]corev3.EntityConfig, wConfigs.Len())
	if err := wConfigs.UnwrapInto(&configs); err != nil {
		return nil, &store.ErrDecode{Err: err, Key: etcdstore.StoreKey(configReq)}
	}

	// Fetch the entity states for each entity with an entity config
	stateRequests := []storev2.ResourceRequest{}
	for _, config := range configs {
		req := storev2.ResourceRequest{
			Namespace:   namespace,
			Name:        config.Metadata.Name,
			Context:     ctx,
			StoreName:   new(corev3.EntityState).StoreName(),
			UsePostgres: true,
		}
		stateRequests = append(stateRequests, req)
	}
	wStates, err := e.store.GetMultiple(ctx, stateRequests)
	if err != nil {
		return nil, err
	}

	// Create a mapping of unique resources (name & namespace) to entity configs
	mappedStates := map[uniqueResource]*corev3.EntityState{}
	for req, wState := range wStates {
		var state corev3.EntityState
		if err := wState.UnwrapInto(&state); err != nil {
			return nil, &store.ErrDecode{Err: err, Key: etcdstore.StoreKey(req)}
		}
		res := uniqueResource{
			Name:      state.Metadata.Name,
			Namespace: state.Metadata.Namespace,
		}
		if _, ok := mappedStates[res]; !ok {
			mappedStates[res] = &state
		} else {
			logger.WithFields(logrus.Fields{
				"name":      state.GetMetadata().GetName(),
				"namespace": state.GetMetadata().GetNamespace(),
			}).Errorf("more than one entity states share the same name & namespace")
		}
	}

	return entitiesFromConfigsAndStates(configs, mappedStates)
}

// Create a list of corev2.Entity values from corev3 configs & states
func entitiesFromConfigsAndStates(configs []corev3.EntityConfig, states map[uniqueResource]*corev3.EntityState) ([]*corev2.Entity, error) {
	entities := []*corev2.Entity{}
	for _, config := range configs {
		res := uniqueResource{
			Name:      config.Metadata.Name,
			Namespace: config.Metadata.Namespace,
		}
		if state, ok := states[res]; ok {
			entity, err := corev3.V3EntityToV2(&config, state)
			if err != nil {
				return nil, &store.ErrNotValid{Err: err}
			}
			entities = append(entities, entity)
		} else {
			// there is a config without a corresponding state, create anyways
			entities = append(entities, entityFromConfigOnly(&config))
		}
	}
	return entities, nil
}

func entityFromConfigOnly(config *corev3.EntityConfig) *corev2.Entity {
	state := corev3.NewEntityState(config.Metadata.Namespace, config.Metadata.Name)
	entity, _ := corev3.V3EntityToV2(config, state)
	return entity
}

// GetEntityConfigByName returns an entity config using the given name and the
// namespace stored in ctx. The resulting entity config is nil if none was
// found.
func (e *EntityStore) GetEntityConfigByName(ctx context.Context, name string) (*corev3.EntityConfig, error) {
	if name == "" {
		return nil, &store.ErrNotValid{Err: errors.New("must specify name")}
	}
	cfg := &corev3.EntityConfig{
		Metadata: &corev2.ObjectMeta{
			Name:      name,
			Namespace: corev2.ContextNamespace(ctx),
		},
	}
	req := storev2.NewResourceRequestFromResource(ctx, cfg)
	req.UsePostgres = true
	wrapper, err := e.store.Get(req)
	if err != nil {
		if _, ok := err.(*store.ErrNotFound); ok {
			return nil, nil
		}
		return nil, err
	}

	if err := wrapper.UnwrapInto(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// GetEntityStateByName returns an entity state using the given name and the
// namespace stored in ctx. The resulting entity state is nil if none was
// found.
func (e *EntityStore) GetEntityStateByName(ctx context.Context, name string) (*corev3.EntityState, error) {
	if name == "" {
		return nil, &store.ErrNotValid{Err: errors.New("must specify name")}
	}
	state := &corev3.EntityState{
		Metadata: &corev2.ObjectMeta{
			Name:      name,
			Namespace: corev2.ContextNamespace(ctx),
		},
	}
	req := storev2.NewResourceRequestFromResource(ctx, state)
	req.UsePostgres = true
	wrapper, err := e.store.Get(req)
	if err != nil {
		if _, ok := err.(*store.ErrNotFound); ok {
			return nil, nil
		}
		return nil, err
	}

	if err := wrapper.UnwrapInto(state); err != nil {
		return nil, err
	}
	return state, nil
}

// GetEntityByName returns an entity using the given name and the namespace stored
// in ctx. The resulting entity is nil if none was found.
func (e *EntityStore) GetEntityByName(ctx context.Context, name string) (*corev2.Entity, error) {
	cfg, err := e.GetEntityConfigByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("error fetching entity config: %w", err)
	}
	if cfg == nil {
		return nil, nil
	}
	state, err := e.GetEntityStateByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("error fetching entity state: %w", err)
	}
	if state == nil {
		return entityFromConfigOnly(cfg), nil
	}
	return corev3.V3EntityToV2(cfg, state)
}

// UpdateEntityConfig creates or updates a given entity config.
func (e *EntityStore) UpdateEntityConfig(ctx context.Context, cfg *corev3.EntityConfig) error {
	if cfg.Metadata.Namespace == "" {
		cfg.Metadata.Namespace = corev2.ContextNamespace(ctx)
	}
	req := storev2.NewResourceRequestFromResource(ctx, cfg)
	req.UsePostgres = true
	wrappedConfig, err := storev2.WrapResource(cfg)
	if err != nil {
		return &store.ErrEncode{Err: err}
	}
	if err := e.store.CreateOrUpdate(req, wrappedConfig); err != nil {
		return err
	}
	return nil
}

// UpdateEntityState creates or updates a given entity state.
func (e *EntityStore) UpdateEntityState(ctx context.Context, state *corev3.EntityState) error {
	if state.Metadata.Namespace == "" {
		state.Metadata.Namespace = corev2.ContextNamespace(ctx)
	}
	req := storev2.NewResourceRequestFromResource(ctx, state)
	req.UsePostgres = true
	wrappedState, err := storev2.WrapResource(state)
	if err != nil {
		return &store.ErrEncode{Err: err}
	}
	if err := e.store.CreateOrUpdate(req, wrappedState); err != nil {
		return err
	}
	return nil
}

// UpdateEntity creates or updates a given entity.
func (e *EntityStore) UpdateEntity(ctx context.Context, entity *corev2.Entity) error {
	namespace := entity.Namespace
	if namespace == "" {
		namespace = corev2.ContextNamespace(ctx)
	}

	cfg, state := corev3.V2EntityToV3(entity)
	cfg.Metadata.Namespace = namespace
	state.Metadata.Namespace = namespace

	if err := e.UpdateEntityConfig(ctx, cfg); err != nil {
		return fmt.Errorf("error updating entity config: %w", err)
	}
	if err := e.UpdateEntityState(ctx, state); err != nil {
		return fmt.Errorf("error updating entity state: %w", err)
	}
	return nil
}

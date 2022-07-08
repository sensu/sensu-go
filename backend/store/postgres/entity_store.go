package postgres

import (
	"context"
	"errors"
	"sync"

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
	stateReq := storev2.NewResourceRequestFromResource(state)
	stateReq.UsePostgres = true
	configReq := storev2.NewResourceRequestFromResource(config)

	if err := e.store.Delete(ctx, configReq); err != nil {
		if _, ok := err.(*store.ErrNotFound); !ok {
			return err
		}
	}
	if err := e.store.Delete(ctx, stateReq); err != nil {
		if _, ok := err.(*store.ErrNotFound); ok {
			return nil
		}
		return err
	}
	return nil
}

// DeleteEntityByName deletes an entity using the given name and the
// namespa	ce stored in ctx.
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
	stateReq := storev2.NewResourceRequestFromResource(state)
	stateReq.UsePostgres = true
	configReq := storev2.NewResourceRequestFromResource(config)
	if err := e.store.Delete(ctx, configReq); err != nil {
		if _, ok := err.(*store.ErrNotFound); !ok {
			return err
		}
	}
	if err := e.store.Delete(ctx, stateReq); err != nil {
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
	var ec corev3.EntityConfig
	configReq := storev2.ResourceRequest{
		Namespace:   namespace,
		Type:        "EntityConfig",
		APIVersion:  "core/v3",
		StoreName:   ec.StoreName(),
		UsePostgres: true,
	}
	if pred.Ordering == corev2.EntitySortName {
		configReq.SortOrder = storev2.SortAscend
		if pred.Descending {
			configReq.SortOrder = storev2.SortDescend
		}
	}

	wConfigs, err := e.store.List(ctx, configReq, pred)
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
			Type:        "EntityState",
			APIVersion:  "core/v3",
			Namespace:   namespace,
			Name:        config.Metadata.Name,
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

// GetEntityByName returns an entity using the given name and the namespace stored
// in ctx. The resulting entity is nil if none was found.
func (e *EntityStore) GetEntityByName(ctx context.Context, name string) (*corev2.Entity, error) {
	if name == "" {
		return nil, &store.ErrNotValid{Err: errors.New("must specify name")}
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
	stateReq := storev2.NewResourceRequestFromResource(state)
	stateReq.UsePostgres = true
	configReq := storev2.NewResourceRequestFromResource(config)
	var wg sync.WaitGroup
	var stateErr error
	wg.Add(1)
	go func() {
		defer wg.Done()
		var wrapper storev2.Wrapper
		wrapper, stateErr = e.store.Get(ctx, stateReq)
		if stateErr != nil {
			return
		}
		stateErr = wrapper.UnwrapInto(state)
	}()

	wrapper, err := e.store.Get(ctx, configReq)
	if err != nil {
		if _, ok := err.(*store.ErrNotFound); ok {
			return nil, nil
		}
		return nil, err
	}

	if err := wrapper.UnwrapInto(config); err != nil {
		return nil, err
	}

	wg.Wait()

	if stateErr != nil {
		if _, ok := stateErr.(*store.ErrNotFound); ok {
			return entityFromConfigOnly(config), nil
		}
		return nil, stateErr
	}

	return corev3.V3EntityToV2(config, state)
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
	stateReq := storev2.NewResourceRequestFromResource(state)
	stateReq.UsePostgres = true
	configReq := storev2.NewResourceRequestFromResource(cfg)
	wrappedState, err := storev2.WrapResource(state)
	if err != nil {
		return &store.ErrEncode{Err: err}
	}
	wrappedConfig, err := storev2.WrapResource(cfg)
	if err != nil {
		return &store.ErrEncode{Err: err}
	}
	if err := e.store.CreateOrUpdate(ctx, configReq, wrappedConfig); err != nil {
		return err
	}
	if err := e.store.CreateOrUpdate(ctx, stateReq, wrappedState); err != nil {
		return err
	}
	return nil
}

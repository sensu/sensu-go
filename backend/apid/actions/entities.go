package actions

import (
	"context"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/backend/store/v2/wrap"
)

// EntityController exposes actions in which a viewer can perform.
type EntityController struct {
	store   store.EntityStore
	storev2 storev2.Interface
}

// NewEntityController returns new EntityController
func NewEntityController(store store.EntityStore, storev2 storev2.Interface) EntityController {
	return EntityController{
		store:   store,
		storev2: storev2,
	}
}

// Find returns resource associated with given parameters if available to the
// viewer.
func (c EntityController) Find(ctx context.Context, id string) (*corev2.Entity, error) {
	// Fetch from store
	result, serr := c.store.GetEntityByName(ctx, id)
	if serr != nil {
		return nil, NewError(InternalErr, serr)
	}
	if result == nil {
		return nil, NewErrorf(NotFound)
	}

	return result, nil
}

// List returns resources available to the viewer.
func (c EntityController) List(ctx context.Context, pred *store.SelectionPredicate) ([]corev2.Resource, error) {
	// Fetch from store
	results, err := c.store.GetEntities(ctx, pred)
	if err != nil {
		return nil, NewError(InternalErr, err)
	}

	resources := make([]corev2.Resource, len(results))
	for i, v := range results {
		resources[i] = v
	}

	return resources, nil
}

// Create instatiates, validates and persists new resource if viewer has access.
func (c EntityController) Create(ctx context.Context, entity corev2.Entity) error {
	// Check for an already existing resource
	if e, err := c.store.GetEntityByName(ctx, entity.Name); err != nil {
		return NewError(InternalErr, err)
	} else if e != nil {
		return NewErrorf(AlreadyExistsErr)
	}

	// Validate the resource
	if err := entity.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Persist the resource in the store
	if err := c.store.UpdateEntity(ctx, &entity); err != nil {
		return NewError(InternalErr, err)
	}
	return nil
}

// CreateOrReplace creates or replaces an entity. It returns an error if the
// provided entity is invalid, the user doesn't have permissions to create or
// update the entity, or if an internal error is returned from the store.
func (c EntityController) CreateOrReplace(ctx context.Context, entity corev2.Entity) error {
	if err := entity.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// TODO(ccressent): add blurb here to explain why we make that exception.
	if entity.EntityClass == corev2.EntityProxyClass {
		if serr := c.store.UpdateEntity(ctx, &entity); serr != nil {
			return NewError(InternalErr, serr)
		}
	} else {
		config, _ := corev3.V2EntityToV3(&entity)
		req := storev2.NewResourceRequestFromResource(ctx, config)

		wConfig, err := wrap.Resource(config)
		if err != nil {
			return err
		}

		if err := c.storev2.CreateOrUpdate(req, wConfig); err != nil {
			return err
		}
	}

	return nil
}

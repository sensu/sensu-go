package actions

import (
	"context"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

var mutatorUpdateFields = []string{
	"Command",
	"Timeout",
	"EnvVars",
	"RuntimeAssets",
}

// MutatorController allows querying mutators in bulk or by name.
type MutatorController struct {
	store store.MutatorStore
}

// NewMutatorController creates a new MutatorController backed by store.
func NewMutatorController(store store.MutatorStore) MutatorController {
	return MutatorController{
		store: store,
	}
}

// Create creates a new Mutator resource.
// It returns non-nil error if the new mutator is invalid, update permissions
// do not exist, or an internal error occurs while updating the underlying
// Store.
func (c MutatorController) Create(ctx context.Context, mut types.Mutator) error {
	// Adjust context
	ctx = addOrgEnvToContext(ctx, &mut)

	// Check for existing
	if m, err := c.store.GetMutatorByName(ctx, mut.Name); err != nil {
		return NewError(InternalErr, err)
	} else if m != nil {
		return NewErrorf(AlreadyExistsErr, mut.Name)
	}

	// Validate
	if err := mut.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Persist
	if err := c.store.UpdateMutator(ctx, &mut); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

// CreateOrReplace creates or replaces a Mutator resource.
// It returns non-nil error if the mutator is invalid, update permissions
// do not exist, or an internal error occurs while updating the underlying
// Store.
func (c MutatorController) CreateOrReplace(ctx context.Context, mut types.Mutator) error {
	// Adjust context
	ctx = addOrgEnvToContext(ctx, &mut)

	// Validate
	if err := mut.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Persist
	if err := c.store.UpdateMutator(ctx, &mut); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

// List returns mutators
func (c MutatorController) List(ctx context.Context, pred *store.SelectionPredicate) ([]corev2.Resource, error) {
	// Fetch from store
	results, err := c.store.GetMutators(ctx, pred)
	if err != nil {
		return nil, NewError(InternalErr, err)
	}

	resources := make([]corev2.Resource, len(results))
	for i, v := range results {
		resources[i] = corev2.Resource(v)
	}

	return resources, nil
}

// Destroy destroys the named Mutator.
// It returns non-nil error if the params are invalid, delete permissions
// do not exist, or an internal error occurs while updating the underlying
// Store.
func (c MutatorController) Destroy(ctx context.Context, name string) error {
	// Validate parameters
	if name == "" {
		return NewErrorf(InvalidArgument, "name is undefined")
	}

	// Fetch from store
	mut, err := c.store.GetMutatorByName(ctx, name)
	if err != nil {
		return NewError(InternalErr, err)
	}
	if mut == nil {
		return NewErrorf(NotFound, name)
	}

	// Remove from store
	if err := c.store.DeleteMutatorByName(ctx, mut.Name); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

// Find returns resource associated with given parameters if available to the
// viewer.
// It returns non-nil error if the params are invalid, read permissions
// do not exist, or an internal error occurs while reading the underlying
// Store.
func (c MutatorController) Find(ctx context.Context, name string) (*types.Mutator, error) {
	result, err := c.store.GetMutatorByName(ctx, name)
	if err != nil {
		return nil, NewErrorf(InternalErr, err)
	}

	if result == nil {
		return nil, NewErrorf(NotFound)
	}

	return result, nil
}

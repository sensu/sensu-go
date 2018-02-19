package actions

import (
	"context"

	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

var mutatorUpdateFields = []string{
	"Command",
	"Timeout",
	"EnvVars",
}

// MutatorController allows querying mutators in bulk or by name.
type MutatorController struct {
	Store  store.MutatorStore
	Policy authorization.MutatorPolicy
}

// NewMutatorController creates a new MutatorController backed by store.
func NewMutatorController(store store.MutatorStore) MutatorController {
	return MutatorController{
		Store:  store,
		Policy: authorization.Mutators,
	}
}

// Create creates a new Mutator resource.
// It returns non-nil error if the new mutator is invalid, update permissions
// do not exist, or an internal error occurs while updating the underlying
// Store.
func (c MutatorController) Create(ctx context.Context, mut types.Mutator) error {
	// Adjust context
	ctx = addOrgEnvToContext(ctx, &mut)
	policy := c.Policy.WithContext(ctx)

	// Check for existing
	if m, err := c.Store.GetMutatorByName(ctx, mut.Name); err != nil {
		return NewError(InternalErr, err)
	} else if m != nil {
		return NewErrorf(AlreadyExistsErr, mut.Name)
	}

	// Verify permissions
	if ok := policy.CanCreate(&mut); !ok {
		return NewErrorf(PermissionDenied, "create")
	}

	// Validate
	if err := mut.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Persist
	if err := c.Store.UpdateMutator(ctx, &mut); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

// Update updates a mutator.
// It returns non-nil error if the new mutator is invalid, create permissions
// do not exist, or an internal error occurs while updating the underlying
// Store.
func (c MutatorController) Update(ctx context.Context, delta types.Mutator) error {
	// Adjust context
	ctx = addOrgEnvToContext(ctx, &delta)
	policy := c.Policy.WithContext(ctx)

	// Check for existing
	mut, err := c.Store.GetMutatorByName(ctx, delta.Name)
	if err != nil {
		return NewError(InternalErr, err)
	} else if mut == nil {
		return NewErrorf(NotFound, delta.Name)
	}

	// Verify viewer can make change
	if ok := policy.CanUpdate(mut); !ok {
		return NewErrorf(PermissionDenied, "update")
	}

	// Update
	if err := mut.Update(&delta, mutatorUpdateFields...); err != nil {
		return NewError(InternalErr, err)
	}

	// Validate
	if err := mut.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Persist
	if err := c.Store.UpdateMutator(ctx, mut); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

// Query returns resources available to the viewer filter by given params.
// It returns non-nil error if the params are invalid, read permissions
// do not exist, or an internal error occurs while reading the underlying
// Store.
func (c MutatorController) Query(ctx context.Context) ([]*types.Mutator, error) {
	policy := c.Policy.WithContext(ctx)

	// Fetch from store
	mutators, err := c.Store.GetMutators(ctx)
	if err != nil {
		return nil, NewError(InternalErr, err)
	}

	result := make([]*types.Mutator, 0, len(mutators))

	// Filter out those resources the viewer does not have access to view.
	for _, m := range mutators {
		if ok := policy.CanRead(m); ok {
			result = append(result, m)
		}
	}

	return result, nil
}

// Destroy destroys the named Mutator.
// It returns non-nil error if the params are invalid, delete permissions
// do not exist, or an internal error occurs while updating the underlying
// Store.
func (c MutatorController) Destroy(ctx context.Context, name string) error {
	policy := c.Policy.WithContext(ctx)

	// Verify permissions
	if ok := policy.CanDelete(); !ok {
		return NewErrorf(PermissionDenied, "delete")
	}

	// Validate parameters
	if name == "" {
		return NewErrorf(InvalidArgument, "name is undefined")
	}

	// Fetch from store
	mut, err := c.Store.GetMutatorByName(ctx, name)
	if err != nil {
		return NewError(InternalErr, err)
	}
	if mut == nil {
		return NewErrorf(NotFound, name)
	}

	// Remove from store
	if err := c.Store.DeleteMutatorByName(ctx, mut.Name); err != nil {
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
	result, err := c.Store.GetMutatorByName(ctx, name)
	if err != nil {
		return nil, NewErrorf(InternalErr, err)
	}

	if result == nil {
		return nil, NewErrorf(NotFound)
	}

	policy := c.Policy.WithContext(ctx)

	if !policy.CanRead(result) {
		return nil, NewErrorf(NotFound)
	}

	return result, nil
}

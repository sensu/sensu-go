package actions

import (
	"context"

	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

var envUpdateFields = []string{
	"Description",
}

// EnvironmentController allows querying Environments in bulk or by name.
type EnvironmentController struct {
	Policy authorization.EnvironmentPolicy
	Store  store.EnvironmentStore
}

// NewEnvironmentController creates a new EnvironmentController backed by store.
func NewEnvironmentController(store store.EnvironmentStore) EnvironmentController {
	return EnvironmentController{
		Store:  store,
		Policy: authorization.EnvironmentPolicy{},
	}
}

// Create creates a new Environment resource.
// It returns non-nil error if the new Filter is invalid, update permissions
// do not exist, or an internal error occurs while updating the underlying
// store.
func (c EnvironmentController) Create(ctx context.Context, env types.Environment) error {
	// Adjust context
	ctx = addOrgEnvToContext(ctx, &env)
	policy := c.Policy.WithContext(ctx)

	// Validate
	if err := env.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Check for existing
	if e, err := c.Store.GetEnvironment(ctx, env.Organization, env.Name); err != nil {
		return NewError(InternalErr, err)
	} else if e != nil {
		return NewErrorf(AlreadyExistsErr, env.Name)
	}

	// Verify permissions
	if ok := policy.CanCreate(&env); !ok {
		return NewErrorf(PermissionDenied, "create")
	}

	// Persist
	if err := c.Store.UpdateEnvironment(ctx, &env); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

// Update updates a Filter.
// It returns non-nil error if the new Filter is invalid, create permissions
// do not exist, or an internal error occurs while updating the underlying
// store.
func (c EnvironmentController) Update(ctx context.Context, delta types.Environment) error {
	if err := delta.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}
	// Adjust context
	ctx = addOrgEnvToContext(ctx, &delta)
	policy := c.Policy.WithContext(ctx)

	// Check for existing
	env, err := c.Store.GetEnvironment(ctx, delta.Organization, delta.Name)
	if err != nil {
		return NewError(InternalErr, err)
	} else if env == nil {
		return NewErrorf(NotFound, delta.Name)
	}

	// Verify viewer can make change
	if ok := policy.CanUpdate(env); !ok {
		return NewErrorf(PermissionDenied, "update")
	}

	// Update
	if err := env.Update(&delta, envUpdateFields...); err != nil {
		return NewError(InternalErr, err)
	}

	// Validate
	if err := env.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Persist
	if err := c.Store.UpdateEnvironment(ctx, env); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

// Query returns resources available to the viewer filter by given params.
// It returns non-nil error if the params are invalid, read permissions
// do not exist, or an internal error occurs while reading the underlying
// store.
func (c EnvironmentController) Query(ctx context.Context, org string) ([]*types.Environment, error) {
	policy := c.Policy.WithContext(ctx)

	// Fetch from store
	envs, err := c.Store.GetEnvironments(ctx, org)
	if err != nil {
		return nil, NewError(InternalErr, err)
	}

	result := make([]*types.Environment, 0, len(envs))

	// Filter out those resources the viewer does not have access to view.
	for _, m := range envs {
		if ok := policy.CanRead(m); ok {
			result = append(result, m)
		}
	}

	return result, nil
}

// Destroy destroys the named Environment.
// It returns non-nil error if the params are invalid, delete permissions
// do not exist, or an internal error occurs while updating the underlying
// Store.
func (c EnvironmentController) Destroy(ctx context.Context, org, name string) error {
	// Validate parameters
	if org == "" {
		return NewErrorf(InvalidArgument, "org is undefined")
	}
	if name == "" {
		return NewErrorf(InvalidArgument, "name is undefined")
	}

	policy := c.Policy.WithContext(ctx)

	// Verify permissions
	if ok := policy.CanDelete(); !ok {
		return NewErrorf(PermissionDenied, "delete")
	}

	// Fetch from store
	env, err := c.Store.GetEnvironment(ctx, org, name)
	if err != nil {
		return NewError(InternalErr, err)
	}
	if env == nil {
		return NewErrorf(NotFound, name)
	}

	// Remove from store
	if err := c.Store.DeleteEnvironment(ctx, env); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

// Find returns resource associated with given parameters if available to the
// viewer.
// It returns non-nil error if the params are invalid, read permissions
// do not exist, or an internal error occurs while reading the underlying
// Store.
func (c EnvironmentController) Find(ctx context.Context, org, name string) (*types.Environment, error) {
	// Find (for mutators) requires a name
	if name == "" {
		return nil, NewErrorf(InvalidArgument, "Find() requires a name")
	}

	result, err := c.Store.GetEnvironment(ctx, org, name)
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

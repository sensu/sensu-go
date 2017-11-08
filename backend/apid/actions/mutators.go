package actions

import (
	"context"

	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

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
func (c MutatorController) Create(ctx context.Context, mut types.Mutator) error {
	ctx = addOrgEnvToContext(ctx, &mut)
	policy := c.Policy.WithContext(ctx)

	// Check for existing
	if m, err := c.Store.GetMutatorByName(ctx, mut.Name); err != nil {
		return NewError(InternalErr, err)
	} else if m != nil {
		return NewErrorf(AlreadyExistsErr)
	}

	// Verify permissions
	if ok := policy.CanCreate(&mut); !ok {
		return NewErrorf(PermissionDenied)
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

// Query returns resources available to the viewer filter by given params.
func (c MutatorController) Query(ctx context.Context, params QueryParams) ([]interface{}, error) {
	policy := c.Policy.WithContext(ctx)

	// Fetch from store
	mutators, err := c.Store.GetMutators(ctx)
	if err != nil {
		return nil, NewError(InternalErr, err)
	}

	result := make([]interface{}, 0, len(mutators))

	// Filter out those resources the viewer does not have access to view.
	for _, m := range mutators {
		if ok := policy.CanRead(m); ok {
			result = append(result, m)
		}
	}

	return result, nil
}

// Find returns resource associated with given parameters if available to the
// viewer.
func (c MutatorController) Find(ctx context.Context, params QueryParams) (interface{}, error) {
	// Find (for mutators) requires a name
	name := params["name"]
	if name == "" {
		return nil, NewErrorf(InvalidArgument, "Find() requires a name")
	}

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

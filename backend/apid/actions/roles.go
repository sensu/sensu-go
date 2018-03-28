package actions

import (
	"context"

	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// roleUpdateFields refers to fields a viewer may update
var roleUpdateFields = []string{"Rules"}

// RoleController exposes actions in which a viewer can perform.
type RoleController struct {
	Store  store.RBACStore
	Policy authorization.RolePolicy
}

// NewRoleController returns new RoleController
func NewRoleController(store store.RBACStore) RoleController {
	return RoleController{
		Store:  store,
		Policy: authorization.Roles,
	}
}

// Query returns resources available to the viewer filter by given params.
func (a RoleController) Query(ctx context.Context) ([]*types.Role, error) {
	// Fetch from store
	results, serr := a.Store.GetRoles(ctx)
	if serr != nil {
		return nil, NewError(InternalErr, serr)
	}

	// Filter out those resources the viewer does not have access to view.
	abilities := a.Policy.WithContext(ctx)
	for i := 0; i < len(results); i++ {
		if !abilities.CanRead(results[i]) {
			results = append(results[:i], results[i+1:]...)
			i--
		}
	}

	return results, nil
}

// Find returns resource associated with given parameters if available to the
// viewer.
func (a RoleController) Find(ctx context.Context, name string) (*types.Role, error) {
	// Fetch from store
	result, serr := a.findRole(ctx, name)
	if serr != nil {
		return nil, serr
	}

	// Verify role has permission to view
	abilities := a.Policy.WithContext(ctx)
	if result != nil && abilities.CanRead(result) {
		return result, nil
	}

	return nil, NewErrorf(NotFound)
}

// Create creates a new role. It returns an error if the role already exists.
func (a RoleController) Create(ctx context.Context, newRole types.Role) error {
	// Role for existing
	if e, err := a.Store.GetRoleByName(ctx, newRole.Name); err != nil {
		return NewError(InternalErr, err)
	} else if e != nil {
		return NewErrorf(AlreadyExistsErr)
	}

	// Verify viewer can make change
	abilities := a.Policy.WithContext(ctx)
	if yes := abilities.CanCreate(); !yes {
		return NewErrorf(PermissionDenied)
	}

	// Validate
	if err := newRole.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Persist
	if err := a.Store.UpdateRole(ctx, &newRole); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

// CreateOrReplace creates or replaces a role.
func (a RoleController) CreateOrReplace(ctx context.Context, newRole types.Role) error {
	// Verify viewer can make change
	abilities := a.Policy.WithContext(ctx)
	if !(abilities.CanCreate() && abilities.CanUpdate()) {
		return NewErrorf(PermissionDenied)
	}

	// Validate
	if err := newRole.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Persist
	if err := a.Store.UpdateRole(ctx, &newRole); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

// Update validates and persists changes to a resource if viewer has access.
func (a RoleController) Update(ctx context.Context, given types.Role) error {
	return a.findAndUpdateRole(ctx, given.Name, func(role *types.Role) error {
		copyFields(role, &given, roleUpdateFields...)
		return nil
	})
}

// Destroy removes given role from the store.
func (a RoleController) Destroy(ctx context.Context, name string) error {
	// Verify role has permission
	abilities := a.Policy.WithContext(ctx)
	if yes := abilities.CanDelete(); !yes {
		return NewErrorf(PermissionDenied)
	}

	// Fetch from store
	_, err := a.findRole(ctx, name)
	if err != nil {
		return err
	}

	// Remove from store
	if serr := a.Store.DeleteRoleByName(ctx, name); serr != nil {
		return NewError(InternalErr, serr)
	}

	return nil
}

// AddRule adds a given rule to a role
func (a RoleController) AddRule(ctx context.Context, role string, rule types.Rule) error {
	return a.findAndUpdateRole(ctx, role, func(role *types.Role) error {
		var exists bool
		for i, r := range role.Rules {
			if r.Type == rule.Type {
				exists = true
				role.Rules[i] = rule
				break
			}
		}

		if !exists {
			role.Rules = append(role.Rules, rule)
		}
		return nil
	})
}

// RemoveRule removes a given rule to a role
func (a RoleController) RemoveRule(ctx context.Context, role string, rType string) error {
	return a.findAndUpdateRole(ctx, role, func(role *types.Role) error {
		for i, r := range role.Rules {
			if r.Type == rType {
				role.Rules = append(role.Rules[:i], role.Rules[i+1:]...)
				break
			}
		}

		return nil
	})
}

func (a RoleController) findRole(ctx context.Context, name string) (*types.Role, error) {
	result, serr := a.Store.GetRoleByName(ctx, name)
	if serr != nil {
		return nil, NewError(InternalErr, serr)
	} else if result == nil {
		return nil, NewErrorf(NotFound)
	}

	return result, nil
}

func (a RoleController) updateRole(ctx context.Context, role *types.Role) error {
	if err := role.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	if err := a.Store.UpdateRole(ctx, role); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

func (a RoleController) findAndUpdateRole(
	ctx context.Context,
	name string,
	configureFn func(*types.Role) error,
) error {
	// Find
	role, serr := a.findRole(ctx, name)
	if serr != nil {
		return serr
	}

	// Verify viewer can make change
	abilities := a.Policy.WithContext(ctx)
	if yes := abilities.CanUpdate(); !yes {
		return NewErrorf(PermissionDenied)
	}

	// Configure
	if err := configureFn(role); err != nil {
		return err
	}

	// Update
	return a.updateRole(ctx, role)
}

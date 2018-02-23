package actions

import (
	"fmt"
	"strings"

	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"golang.org/x/net/context"
)

// UserController exposes actions in which a viewer can perform.
type UserController struct {
	Store interface {
		store.UserStore
		store.RBACStore
	}
	Policy authorization.UserPolicy
}

// NewUserController returns new UserController
func NewUserController(store store.Store) UserController {
	return UserController{
		Store:  store,
		Policy: authorization.Users,
	}
}

// Query returns resources available to the viewer filter by given params.
func (a UserController) Query(ctx context.Context) ([]*types.User, error) {
	// Fetch from store
	results, serr := a.Store.GetAllUsers()
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
func (a UserController) Find(ctx context.Context, name string) (*types.User, error) {
	// Fetch from store
	result, serr := a.findUser(ctx, name)
	if serr != nil {
		return nil, serr
	}

	// Verify user has permission to view
	abilities := a.Policy.WithContext(ctx)
	if result != nil && abilities.CanRead(result) {
		return result, nil
	}

	return nil, NewErrorf(NotFound)
}

// Create instantiates, validates and persists new resource if viewer has access.
func (a UserController) Create(ctx context.Context, newUser types.User) error {
	// User for existing
	if e, err := a.Store.GetUser(ctx, newUser.Username); err != nil {
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
	if err := newUser.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Validate password
	if err := newUser.ValidatePassword(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Validate roles
	if err := validateRoles(ctx, a.Store, newUser.Roles); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Persist
	if err := a.Store.UpdateUser(&newUser); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

// Update validates and persists changes to a resource if viewer has access.
func (a UserController) Update(ctx context.Context, given types.User) error {
	// Find existing user
	user, serr := a.findUser(ctx, given.Username)
	if serr != nil {
		return serr
	}

	// Setup authorization policy
	abilities := a.Policy.WithContext(ctx)

	// Copy & validate password if given
	if given.Password != "" {
		user.Password = given.Password

		// Verify viewer can make change
		if yes := abilities.CanChangePassword(user); !yes {
			return NewErrorf(
				PermissionDenied,
				"insufficient access to update password",
			)
		}

		// Validate password
		if err := user.ValidatePassword(); err != nil {
			return NewError(InvalidArgument, err)
		}
	}

	// Copy & validate new roles, if given
	if given.Roles != nil {
		if err := configureRoles(ctx, a.Store, &abilities, given.Roles, user); err != nil {
			return err
		}
	}

	// Persist Changes
	return a.updateUser(ctx, user)
}

// Disable disables user identified by given name if viewer has access.
func (a UserController) Disable(ctx context.Context, name string) error {
	// Fetch from store
	result, serr := a.findUser(ctx, name)
	if serr != nil {
		return serr
	}

	// Verify user has permission
	abilities := a.Policy.WithContext(ctx)
	if yes := abilities.CanDelete(result); !yes {
		return NewErrorf(PermissionDenied)
	}

	// Disable
	if !result.Disabled {
		if serr := a.Store.DeleteUser(ctx, result); serr != nil {
			return NewError(InternalErr, serr)
		}
	}

	return nil
}

// Enable disables user identified by given name if viewer has access.
func (a UserController) Enable(ctx context.Context, name string) error {
	abilities := a.Policy.WithContext(ctx)

	// Fetch from store
	result, serr := a.findUser(ctx, name)
	if serr != nil {
		return serr
	}

	// Verify user has permission
	if yes := abilities.CanUpdate(result); !yes {
		return NewErrorf(PermissionDenied)
	}

	// Re-enable
	var err error
	if result.Disabled {
		result.Disabled = false
		err = a.updateUser(ctx, result)
	}

	return err
}

// AddRole adds a given role to a user
func (a UserController) AddRole(ctx context.Context, username string, role string) error {
	return a.findAndUpdateUser(ctx, username, func(user *types.User) error {
		var exists bool
		for _, r := range user.Roles {
			if r == role {
				exists = true
				break
			}
		}

		if !exists {
			newRoles := append(user.Roles, role)
			abilities := a.Policy.WithContext(ctx)
			return configureRoles(ctx, a.Store, &abilities, newRoles, user)
		}

		return nil
	})
}

// RemoveRole adds a given role to a user
func (a UserController) RemoveRole(ctx context.Context, username string, role string) error {
	return a.findAndUpdateUser(ctx, username, func(user *types.User) error {
		newRoles := []string{}
		for _, r := range user.Roles {
			if r != role {
				newRoles = append(newRoles, r)
			}
		}
		user.Roles = newRoles

		// Verify viewer can make change
		abilities := a.Policy.WithContext(ctx)
		if yes := abilities.CanUpdate(user); !yes {
			return NewErrorf(PermissionDenied)
		}

		return nil
	})
}

func (a UserController) findUser(ctx context.Context, name string) (*types.User, error) {
	result, serr := a.Store.GetUser(ctx, name)
	if serr != nil {
		return nil, NewError(InternalErr, serr)
	} else if result == nil {
		return nil, NewErrorf(NotFound)
	}

	return result, nil
}

func (a UserController) updateUser(ctx context.Context, user *types.User) error {
	if err := a.Store.UpdateUser(user); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

func (a UserController) findAndUpdateUser(
	ctx context.Context,
	name string,
	configureFn func(*types.User) error,
) error {
	// Find
	user, serr := a.findUser(ctx, name)
	if serr != nil {
		return serr
	}

	// Configure
	if err := configureFn(user); err != nil {
		return err
	}

	// Update
	return a.updateUser(ctx, user)
}

func configureRoles(
	ctx context.Context,
	store store.RBACStore,
	abilities *authorization.UserPolicy,
	newRoles []string,
	user *types.User,
) error {
	user.Roles = newRoles

	// Verify viewer can make change
	if yes := abilities.CanUpdate(user); !yes {
		return NewErrorf(PermissionDenied)
	}

	// Validate roles
	if err := validateRoles(ctx, store, user.Roles); err != nil {
		return NewError(InvalidArgument, err)
	}

	return nil
}

func validateRoles(ctx context.Context, store store.RBACStore, givenRoles []string) error {
	storedRoles, err := store.GetRoles(ctx)
	if err != nil {
		return err
	}

	missingRoles := []string{}

	for _, givenRole := range givenRoles {
		if present := hasRole(storedRoles, givenRole); !present {
			missingRoles = append(missingRoles, givenRole)
		}
	}

	if len(missingRoles) != 0 {
		message := "not exist and should be created first"
		if len(missingRoles) == 1 {
			message = fmt.Sprintf("given role '%s' does %s", missingRoles[0], message)
		} else {
			message = fmt.Sprintf(
				"given roles '%s' do %s",
				strings.Join(missingRoles, ", "),
				message,
			)
		}
		return fmt.Errorf(message)
	}

	return nil
}

func hasRole(roles []*types.Role, roleName string) bool {
	for _, role := range roles {
		if roleName == role.Name {
			return true
		}
	}
	return false
}

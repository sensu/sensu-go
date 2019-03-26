package actions

import (
	"context"
	"encoding/base64"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/authentication/bcrypt"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// UserController exposes actions in which a viewer can perform.
type UserController struct {
	Store interface {
		store.UserStore
	}
}

// NewUserController returns new UserController
func NewUserController(store store.Store) UserController {
	return UserController{
		Store: store,
	}
}

// Query returns resources available to the viewer filter by given params.
func (a UserController) Query(ctx context.Context) ([]*types.User, string, error) {
	pageSize := corev2.PageSizeFromContext(ctx)
	continueToken := corev2.PageContinueFromContext(ctx)

	// Fetch from store
	results, newContinueToken, serr := a.Store.GetAllUsers(int64(pageSize), continueToken)
	if serr != nil {
		return nil, "", NewError(InternalErr, serr)
	}

	// Encode the continue token with base64url (RFC 4648), without padding
	encodedNewContinueToken := base64.RawURLEncoding.EncodeToString([]byte(newContinueToken))

	return results, encodedNewContinueToken, nil
}

// Find returns resource associated with given parameters if available to the
// viewer.
func (a UserController) Find(ctx context.Context, name string) (*types.User, error) {
	// Fetch from store
	result, serr := a.findUser(ctx, name)
	if serr != nil {
		return nil, serr
	}
	if result == nil {
		return nil, NewErrorf(NotFound)
	}

	return result, nil
}

// Create creates a new user. It returns an error if the user already exists.
func (a UserController) Create(ctx context.Context, newUser types.User) error {
	// Check for existing
	if e, err := a.Store.GetUser(ctx, newUser.Username); err != nil {
		return NewError(InternalErr, err)
	} else if e != nil {
		return NewErrorf(AlreadyExistsErr)
	}

	return a.CreateOrReplace(ctx, newUser)
}

// CreateOrReplace creates or replaces a user.
func (a UserController) CreateOrReplace(ctx context.Context, newUser types.User) error {
	// Validate
	if err := newUser.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Validate password
	if err := newUser.ValidatePassword(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Create password digest
	hash, err := bcrypt.HashPassword(newUser.Password)
	if err != nil {
		return NewError(InternalErr, err)
	}
	newUser.Password = hash

	// Persist
	if err := a.Store.UpdateUser(&newUser); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

// Disable disables user identified by given name if viewer has access.
func (a UserController) Disable(ctx context.Context, name string) error {
	// Fetch from store
	result, serr := a.findUser(ctx, name)
	if serr != nil {
		return serr
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
	// Fetch from store
	result, serr := a.findUser(ctx, name)
	if serr != nil {
		return serr
	}

	// Re-enable
	var err error
	if result.Disabled {
		result.Disabled = false
		err = a.updateUser(ctx, result)
	}

	return err
}

// AddGroup adds a given group to a user
func (a UserController) AddGroup(ctx context.Context, username string, group string) error {
	return a.findAndUpdateUser(ctx, username, func(user *types.User) error {
		var exists bool
		for _, g := range user.Groups {
			if g == group {
				exists = true
				break
			}
		}

		if !exists {
			user.Groups = append(user.Groups, group)
		}

		return nil
	})
}

// RemoveGroup removes a group from a given user
func (a UserController) RemoveGroup(ctx context.Context, username string, group string) error {
	return a.findAndUpdateUser(ctx, username, func(user *types.User) error {
		updatedGroups := []string{}
		for _, g := range user.Groups {
			if g != group {
				updatedGroups = append(updatedGroups, g)
			}
		}

		user.Groups = updatedGroups
		return nil
	})
}

// RemoveAllGroups removes all groups from a given user
func (a UserController) RemoveAllGroups(ctx context.Context, username string) error {
	return a.findAndUpdateUser(ctx, username, func(user *types.User) error {
		user.Groups = []string{}
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

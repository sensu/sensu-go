package actions

import (
	"context"
	"errors"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/authentication/bcrypt"
	"github.com/sensu/sensu-go/backend/store"
)

// UserController exposes actions in which a viewer can perform.
type UserController struct {
	store store.UserStore
}

// NewUserController returns new UserController
func NewUserController(store store.Store) UserController {
	return UserController{
		store: store,
	}
}

// List returns resources available to the viewer filter by given params.
func (a UserController) List(ctx context.Context, pred *store.SelectionPredicate) ([]corev2.Resource, error) {
	// Fetch from store
	users, err := a.store.GetAllUsers(pred)
	if err != nil {
		return nil, NewError(InternalErr, err)
	}

	resources := make([]corev2.Resource, len(users))
	for i, user := range users {
		// Obfuscate the password hashes for now
		user.Password = ""
		user.PasswordHash = ""
		resources[i] = corev2.Resource(user)
	}

	return resources, nil
}

// Get returns resource associated with given parameters if available to the
// viewer.
func (a UserController) Get(ctx context.Context, name string) (*corev2.User, error) {
	// Fetch from store
	user, serr := a.findUser(ctx, name)
	if serr != nil {
		return nil, serr
	}
	if user == nil {
		return nil, NewErrorf(NotFound)
	}

	// Obfuscate the password hashes for now
	user.Password = ""
	user.PasswordHash = ""

	return user, nil
}

// Create creates a new user. It returns an error if the user already exists.
func (a UserController) Create(ctx context.Context, user *corev2.User) error {
	// Check for existing
	if e, err := a.store.GetUser(ctx, user.Username); err != nil {
		return NewError(InternalErr, err)
	} else if e != nil {
		return NewErrorf(AlreadyExistsErr)
	}

	return a.CreateOrReplace(ctx, user)
}

// CreateOrReplace creates or replaces a user.
func (a UserController) CreateOrReplace(ctx context.Context, user *corev2.User) error {
	// Validate
	if err := user.Validate(); err != nil {
		return NewError(InvalidArgument, err)
	}

	// Determine if a hashed and/or cleartext password was provided
	if user.Password != "" && user.PasswordHash != "" {
		// Both the cleartext & hashed passwords were provided, so we need to make
		// sure they match
		if ok := bcrypt.CheckPassword(user.PasswordHash, user.Password); !ok {
			return NewError(
				InvalidArgument,
				errors.New("hashed password does not the match the cleartext password, only one of those should be provided"),
			)
		}
	} else if user.Password != "" {
		// We need to validate the cleartext passsword so it matches our minimal
		// requirements
		if err := user.ValidatePassword(); err != nil {
			return NewError(InvalidArgument, err)
		}

		// Create a hash for this password
		hash, err := bcrypt.HashPassword(user.Password)
		if err != nil {
			return NewError(InternalErr, err)
		}
		user.PasswordHash = hash
	} else if user.PasswordHash == "" {
		return NewError(InvalidArgument, errors.New("a password or its hash is required"))
	}

	// Also add the hash to the password field for backward compatibility
	user.Password = user.PasswordHash

	// Persist
	if err := a.store.UpdateUser(user); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

// Disable disables user identified by given name if viewer has access.
func (a UserController) Disable(ctx context.Context, name string) error {
	// Fetch from store
	user, serr := a.findUser(ctx, name)
	if serr != nil {
		return serr
	}

	if user.Disabled {
		return nil
	}

	user.Disabled = true
	if err := a.store.UpdateUser(user); err != nil {
		return NewError(InternalErr, err)
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
	return a.findAndUpdateUser(ctx, username, func(user *corev2.User) error {
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
	return a.findAndUpdateUser(ctx, username, func(user *corev2.User) error {
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
	return a.findAndUpdateUser(ctx, username, func(user *corev2.User) error {
		user.Groups = []string{}
		return nil
	})
}

func (a UserController) findUser(ctx context.Context, name string) (*corev2.User, error) {
	result, serr := a.store.GetUser(ctx, name)
	if serr != nil {
		return nil, NewError(InternalErr, serr)
	} else if result == nil {
		return nil, NewErrorf(NotFound)
	}

	return result, nil
}

func (a UserController) updateUser(ctx context.Context, user *corev2.User) error {
	if err := a.store.UpdateUser(user); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

func (a UserController) findAndUpdateUser(
	ctx context.Context,
	name string,
	configureFn func(*corev2.User) error,
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

// AuthenticateUser attempts to authenticate an internal user
func (a UserController) AuthenticateUser(ctx context.Context, username, password string) (*corev2.User, error) {
	return a.store.AuthenticateUser(ctx, username, password)
}

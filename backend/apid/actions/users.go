package actions

import (
	"context"
	"errors"
	"fmt"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/authentication/bcrypt"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

// UserController exposes actions in which a viewer can perform.
type UserController struct {
	store storev2.Interface
}

// NewUserController returns new UserController
func NewUserController(store storev2.Interface) UserController {
	return UserController{
		store: store,
	}
}

// List returns resources available to the viewer filter by given params.
func (a UserController) List(ctx context.Context, pred *store.SelectionPredicate) ([]corev3.Resource, error) {
	// Fetch from store
	req := storev2.NewResourceRequestFromResource(new(corev2.User))
	var users []*corev2.User
	list, err := a.store.List(ctx, req, nil)
	if err != nil {
		switch err := err.(type) {
		case *store.ErrNotFound:
			return nil, NewErrorf(NotFound)
		default:
			return nil, NewError(InternalErr, err)
		}
	}
	if err := list.UnwrapInto(&users); err != nil {
		return nil, NewError(InternalErr, err)
	}

	resources := make([]corev3.Resource, len(users))
	for i, user := range users {
		// Obfuscate the password hashes for now
		user.Password = ""
		user.PasswordHash = ""
		resources[i] = corev3.Resource(user)
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
	req := storev2.NewResourceRequestFromResource(user)
	wrapper, err := storev2.WrapResource(user)
	if err != nil {
		return NewError(InvalidArgument, err)
	}
	if err := a.store.CreateIfNotExists(ctx, req, wrapper); err != nil {
		switch err := err.(type) {
		case *store.ErrAlreadyExists:
			return NewErrorf(AlreadyExistsErr)
		default:
			return NewError(InternalErr, err)
		}
	}
	return nil
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
	req := storev2.NewResourceRequestFromResource(user)
	wrapper, err := storev2.WrapResource(user)
	if err != nil {
		return NewError(InvalidArgument, err)
	}
	if err := a.store.CreateOrUpdate(ctx, req, wrapper); err != nil {
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

	req := storev2.NewResourceRequestFromResource(user)
	wrapper, err := storev2.WrapResource(user)
	if err != nil {
		return NewError(InvalidArgument, err)
	}

	if err := a.store.UpdateIfExists(ctx, req, wrapper); err != nil {
		switch err.(type) {
		case *store.ErrAlreadyExists:
			return NewError(AlreadyExistsErr, err)
		default:
			return NewError(InternalErr, err)
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
	user := &corev2.User{
		Username: name,
	}
	req := storev2.NewResourceRequestFromResource(user)
	wrapper, err := a.store.Get(ctx, req)
	if err != nil {
		switch err := err.(type) {
		case *store.ErrNotFound:
			return nil, NewErrorf(NotFound)
		default:
			return nil, NewError(InternalErr, err)
		}
	}
	if err := wrapper.UnwrapInto(user); err != nil {
		return nil, NewError(InternalErr, err)
	}
	return user, nil
}

func (a UserController) updateUser(ctx context.Context, user *corev2.User) error {
	req := storev2.NewResourceRequestFromResource(user)
	wrapper, err := storev2.WrapResource(user)
	if err != nil {
		return err
	}
	if err := a.store.CreateOrUpdate(ctx, req, wrapper); err != nil {
		switch err := err.(type) {
		case *store.ErrNotFound:
			return NewErrorf(NotFound)
		default:
			return NewError(InternalErr, err)
		}
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
	user, err := a.findUser(ctx, username)
	if err != nil {
		return nil, err
	}
	if user.Disabled {
		return nil, &store.ErrNotValid{Err: fmt.Errorf("user %s is disabled", username)}
	}

	// Check if we have an explicitly hashed password, otherwise fallback to the
	// password field for backward compatiblility
	passwordHash := user.PasswordHash
	if passwordHash == "" {
		passwordHash = user.Password
	}
	ok := bcrypt.CheckPassword(passwordHash, password)
	if !ok {
		return nil, &store.ErrNotValid{Err: fmt.Errorf("wrong password for user %s", username)}
	}

	return user, nil
}

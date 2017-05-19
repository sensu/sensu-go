package basic

import (
	"errors"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"golang.org/x/crypto/bcrypt"
)

const (
	name = "basic"
)

// Basic represents the HTTP basic authentication provider
type Basic struct {
	Store store.Store
}

// Authenticate tries to authenticate the provided username & password combination
func (b *Basic) Authenticate(username, password string) (*types.User, error) {
	return nil, errors.New("Authentication is disabled")
}

// CreateUser adds a new user through the store
func (b *Basic) CreateUser(user *types.User) error {
	hash, err := hashPassword(user.Password)
	if err != nil {
		return err
	}
	user.Password = hash

	return b.Store.CreateUser(user)
}

// Name returns the provider name
func (b *Basic) Name() string {
	return name
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

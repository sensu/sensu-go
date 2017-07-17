package basic

import (
	"fmt"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"golang.org/x/crypto/bcrypt"
)

// Basic represents the HTTP basic authentication provider
type Basic struct {
	Store store.Store
}

// Authenticate tries to authenticate the provided username & password combination
func (b *Basic) Authenticate(username, password string) (*types.User, error) {
	user, err := b.Store.GetUser(username)
	if err != nil {
		return nil, fmt.Errorf("User %s does not exist", username)
	}

	if user.Disabled {
		return nil, fmt.Errorf("User %s is disabled", username)
	}

	ok := checkPassword(user.Password, password)
	if !ok {
		return nil, fmt.Errorf("Wrong password for user %s", username)
	}

	return user, nil
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

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func checkPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

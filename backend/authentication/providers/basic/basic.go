package basic

import (
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"golang.org/x/crypto/bcrypt"
)

// Basic represents the HTTP basic authentication provider
type Basic struct {
	Store store.Store
}

// CreateUser adds a new user through the store
func (b *Basic) CreateUser(user *types.User) error {
	hash, err := hashPassword(user.Password)
	if err != nil {
		return err
	}
	user.Password = hash

	return b.Store.UpdateUser(user)
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

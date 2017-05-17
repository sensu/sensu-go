package basic

import (
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// Basic represents the HTTP basic authentication provider
type Basic struct {
	Store store.Store
}

// CreateUser adds a new user through the store
func (b *Basic) CreateUser(user *types.User) error {
	return b.Store.UpdateUser(user)
}

package basic

import (
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

type Basic struct {
	Store store.Store
}

func (b *Basic) CreateUser(user *types.User) error {
	return b.Store.UpdateUser(user)
}

package authentication

import "github.com/sensu/sensu-go/types"

type Provider interface {
	CreateUser(user *types.User) error
}

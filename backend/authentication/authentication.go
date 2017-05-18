package authentication

import "github.com/sensu/sensu-go/types"

// Provider represents an authenticated provider
type Provider interface {
	CreateUser(user *types.User) error
}

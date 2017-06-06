package authentication

import (
	"github.com/sensu/sensu-go/types"
)

// Provider represents an authenticated provider
type Provider interface {
	Authenticate(string, string) (*types.User, error)
	AuthEnabled() bool
	CreateUser(*types.User) error
}

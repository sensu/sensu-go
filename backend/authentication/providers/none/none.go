package none

import (
	"errors"

	"github.com/sensu/sensu-go/types"
)

const (
	name = "none"
)

// None represents disabled authentication
type None struct{}

// Authenticate tries to authenticate the provided username & password combination
func (n *None) Authenticate(username, password string) (*types.User, error) {
	return nil, errors.New("Authentication is disabled")
}

// CreateUser adds a new user through the store
func (n *None) CreateUser(user *types.User) error {
	return errors.New("Authentication is disabled")
}

// Name returns the provider name
func (n *None) Name() string {
	return name
}

package none

import (
	"errors"

	"github.com/sensu/sensu-go/types"
)

// None represents disabled authentication
type None struct{}

// CreateUser adds a new user through the store
func (n *None) CreateUser(user *types.User) error {
	return errors.New("Authentication is disabled")
}

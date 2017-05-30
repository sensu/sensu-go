package authentication

import (
	"github.com/Sirupsen/logrus"
	"github.com/sensu/sensu-go/types"
)

var logger = logrus.WithFields(logrus.Fields{
	"component": "authentication",
})

// Provider represents an authenticated provider
type Provider interface {
	Authenticate(string, string) (*types.User, error)
	AuthEnabled() bool
	CreateUser(*types.User) error
}

package mockauthorizer

import (
	"context"

	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/stretchr/testify/mock"
)

// Authorizer mocks an authorization provider
type Authorizer struct {
	mock.Mock
}

// Authorize ...
func (a *Authorizer) Authorize(ctx context.Context, attrs *authorization.Attributes) (bool, error) {
	args := a.Called(ctx, attrs)
	return args.Bool(0), args.Error(1)
}

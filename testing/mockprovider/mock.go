package mockprovider

import (
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/mock"
)

// MockProvider is an authentication provier used for testing. When using the
// MockStore in unit tests, stub out the behavior you wish to test against by
// assigning the appropriate function to the appropriate Func field. If you have
// forgotten to stub a particular function, the program will panic.
type MockProvider struct {
	mock.Mock
}

// Authenticate mocks the authentication
func (m *MockProvider) Authenticate(username, password string) (*types.User, error) {
	args := m.Called()
	return args.Get(0).(*types.User), args.Error(1)
}

// CreateUser mocks the user creation
func (m *MockProvider) CreateUser(user *types.User) error {
	args := m.Called()
	return args.Error(0)
}

// Name mocks the provider name
func (m *MockProvider) Name() string {
	args := m.Called()
	return args.String(0)
}

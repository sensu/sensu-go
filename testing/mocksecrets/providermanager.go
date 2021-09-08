package mocksecrets

import (
	"context"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/secrets"
	"github.com/stretchr/testify/mock"
)

// ProviderManager ...
type ProviderManager struct {
	mock.Mock
}

// AddProvider ...
func (m *ProviderManager) AddProvider(provider secrets.Provider) {
	m.Called(provider)
}

// Providers ...
func (m *ProviderManager) Providers() map[string]secrets.Provider {
	args := m.Called()
	return args.Get(0).(map[string]secrets.Provider)
}

// RemoveProvider ...
func (m *ProviderManager) RemoveProvider(name string) error {
	args := m.Called(name)
	return args.Error(1)
}

// SubSecrets ...
func (m *ProviderManager) SubSecrets(ctx context.Context, secrets []*corev2.Secret) ([]string, error) {
	args := m.Called(ctx, secrets)
	return args.Get(0).([]string), args.Error(1)
}

package secrets

import (
	"fmt"
	"strings"
	"sync"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// A Provider gets all secrets from a secrets provider.
type Provider interface {
	corev2.Resource
	// Get gets the value of the secret associated with the Sensu resource name.
	Get(name string) (secret string, err error)
}

// ProviderManager manages the list of secrets providers.
type ProviderManager struct {
	mu        sync.RWMutex
	providers map[string]Provider
}

// AddProvider adds a provider to the list of configured providers.
func (m *ProviderManager) AddProvider(provider Provider) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Make sure the providers map is not nil
	if m.providers == nil {
		m.providers = map[string]Provider{}
	}

	m.providers[provider.GetObjectMeta().Name] = provider
}

// Providers returns the configured providers.
func (m *ProviderManager) Providers() map[string]Provider {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Create a new map and copy the authenticator providers into it in order to
	// prevent race conditions
	providers := make(map[string]Provider, len(m.providers))
	for k, v := range m.providers {
		providers[k] = v
	}

	return providers
}

// RemoveProvider removes the provider with the given name from the list of
// configued providers.
func (m *ProviderManager) RemoveProvider(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.providers[name]; !ok {
		return fmt.Errorf("provider %q is not configured, could not remove it", name)
	}

	delete(m.providers, name)
	return nil
}

// SubSecrets substitutes all secret tokens with the value of the secret.
func (m *ProviderManager) SubSecrets(command string) ([]string, error) {
	secretVars := []string{}
	// iterate through each argument in the command
	args := strings.Split(command, " ")
	for _, a := range args {
		// if the arg starts with $, find the secret
		if strings.HasPrefix(a, "$") {
			providers := m.Providers()
			// iterate through each secrets provider
			for name, p := range providers {
				// ask the provider to retrieve the secret
				secretKey := strings.TrimLeft(a, "$")
				secretValue, err := p.Get(secretKey)
				if err != nil {
					logger.WithField("provider", name).WithError(err).Error("unable to retrieve secrets from provider")
					return []string{}, err
				}
				if secretValue != "" {
					secretVars = append(secretVars, fmt.Sprintf("%s=%s", secretKey, secretValue))
				}
			}
		}
	}

	return secretVars, nil
}

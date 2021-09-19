package secrets

import (
	"context"
	"fmt"
	"sync"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sirupsen/logrus"
)

// Provider represents an abstracted secrets provider.
type Provider interface {
	corev2.Resource
	// Get gets the value of the secret associated with the secret ID.
	Get(id string) (string, error)
}

// ProviderManagerer represents an abstracted secrets provider manager.
type ProviderManagerer interface {
	AddProvider(Provider)
	Providers() map[string]Provider
	RemoveProvider(string) error
	SubSecrets(context.Context, []*corev2.Secret) ([]string, error)
}

// ProviderManager manages the list of secrets providers.
type ProviderManager struct {
	mu         *sync.RWMutex
	providers  map[string]Provider
	TLSenabled bool
	Getter     Getter
}

// NewProviderManager instantiates a new provider manager.
func NewProviderManager() *ProviderManager {
	return &ProviderManager{
		providers: map[string]Provider{},
		mu:        &sync.RWMutex{},
	}
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

	// Create a new map and copy the secrets providers into it in order to
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
func (m *ProviderManager) SubSecrets(ctx context.Context, secrets []*corev2.Secret) ([]string, error) {
	secretVars := []string{}
	// short circuit the function if there are no secrets
	if len(secrets) == 0 {
		return secretVars, nil
	}

	// Make sure the providers map is not nil
	if m.providers == nil {
		m.providers = map[string]Provider{}
	}
	providers := m.Providers()
	// short circuit the function if there are no secrets providers
	if len(providers) == 0 {
		return secretVars, fmt.Errorf("no secrets providers defined")
	}
	if m.Getter == nil {
		return []string{}, fmt.Errorf("secrets management is not supported")
	}

	// iterate through each secret in the config
	for _, secret := range secrets {
		// get the provider name and secret ID associated with the Sensu secret
		providerName, secretID, err := m.Getter.Get(ctx, secret.Secret)
		if err != nil {
			logger.WithFields(logrus.Fields{
				"provider": providerName,
				"secret":   secret.Secret,
			}).WithError(err).Error("unable to retrieve secret from provider")
			return []string{}, err
		}
		provider := providers[providerName]
		if provider == nil {
			err = fmt.Errorf("provider not found, or not working: %s", providerName)
			logger.WithFields(logrus.Fields{
				"provider": providerName,
				"secret":   secret.Secret,
			}).WithError(err).Error("unable to retrieve secret from provider")
			return []string{}, err
		}
		// ask the provider to retrieve the secret
		secretKey := secret.Name
		secretValue, err := provider.Get(secretID)
		if err != nil {
			logger.WithFields(logrus.Fields{
				"provider": providerName,
				"secretID": secretID,
			}).WithError(err).Error("unable to retrieve secret from provider")
			return []string{}, err
		}
		if secretValue != "" {
			secretVars = append(secretVars, fmt.Sprintf("%s=%s", secretKey, secretValue))
		}
	}

	return secretVars, nil
}

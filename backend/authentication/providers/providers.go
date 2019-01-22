package providers

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/sensu/sensu-go/api/core/v2"
)

// Provider represents an abstracted authentication provider
type Provider interface {
	Authenticate(ctx context.Context, username, password string) (*v2.Claims, error)
	GetName() string
	Refresh(ctx context.Context, providerClaims v2.ProviderClaims) (*v2.Claims, error)
	Type() string
}

// ID returns a unique identifier for a given provider
func ID(p Provider) string {
	return fmt.Sprintf("%s/%s", p.Type(), p.GetName())
}

// Authenticator contains the list of authentication providers
type Authenticator struct {
	mu        sync.RWMutex
	providers map[string]Provider
}

// Authenticate with the configured authentication providers
func (a *Authenticator) Authenticate(ctx context.Context, username, password string) (*v2.Claims, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	for _, provider := range a.providers {
		claims, err := provider.Authenticate(ctx, username, password)
		if err != nil || claims == nil {
			logger.WithError(err).Debugf(
				"could not authenticate with provider %q", provider.Type(),
			)
			continue
		}

		return claims, nil
	}

	return nil, errors.New("authentication failed")
}

// Refresh is called when a new access token is requested with a refresh token.
// The provider should attempt to update the user identity to reflect any changes
// since the access token was last refreshed
func (a *Authenticator) Refresh(ctx context.Context, claims v2.ProviderClaims) (*v2.Claims, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if provider, ok := a.providers[claims.ProviderID]; ok {
		user, err := provider.Refresh(ctx, claims)
		if err != nil {
			return nil, fmt.Errorf(
				"could not refresh user %q with provider %q: %s", claims.UserID, provider.Type(), err,
			)
		}

		return user, nil
	}

	return nil, fmt.Errorf(
		"could not refresh user %q with missing provider ID %q", claims.UserID, claims.ProviderID,
	)
}

// AddProvider adds a provided provider to the list of configured providers
func (a *Authenticator) AddProvider(provider Provider) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Make sure the providers map is not nil
	if a.providers == nil {
		a.providers = map[string]Provider{}
	}

	a.providers[ID(provider)] = provider
}

// Providers returns the configured providers
func (a *Authenticator) Providers() map[string]Provider {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.providers
}

// RemoveProvider removes the provider with the given ID from the list of
// configued providers
func (a *Authenticator) RemoveProvider(id string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if _, ok := a.providers[id]; !ok {
		return fmt.Errorf("provider %q is not configured, could not remove it", id)
	}

	// Make sure at least two providers are enabled so there's still one remaining
	if len(a.providers) == 1 {
		return fmt.Errorf("provider %q is the only provider configured, could not remove it ", id)
	}

	delete(a.providers, id)
	return nil
}

package authentication

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"

	corev2 "github.com/sensu/core/v2"
)

// Authenticator contains the list of authentication providers
type Authenticator struct {
	mu        sync.RWMutex
	providers map[string]corev2.AuthProvider
}

// Authenticate with the configured authentication providers
func (a *Authenticator) Authenticate(ctx context.Context, username, password string) (*corev2.Claims, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// TODO(palourde): The Go runtime randomizes map iteration order so the
	// providers resolution order might vary on each authentication, and
	// consequently provoke weird behavior if the same username/password
	// combination exists in multiple providers.
	for _, provider := range a.providers {
		claims, err := provider.Authenticate(ctx, username, password)
		if err != nil || claims == nil {
			logger.WithError(err).Debugf(
				"could not authenticate with provider %q", provider.Type(),
			)
			continue
		}

		logger.WithFields(logrus.Fields{
			"subject":         claims.Subject,
			"groups":          claims.Groups,
			"provider_id":     claims.Provider.ProviderID,
			"provider_type":   claims.Provider.ProviderType,
			"provider_userid": claims.Provider.UserID,
		}).Info("login successful")
		return claims, nil
	}

	// TODO(palourde): We might want to return a more meaningful and actionnable
	// error message, but we don't want to leak sensitive information.

	logger.WithField("username", username).Error("authentication failed")
	return nil, errors.New("authentication failed")
}

// Refresh is called when a new access token is requested with a refresh token.
// The provider should attempt to update the user identity to reflect any changes
// since the access token was last refreshed
func (a *Authenticator) Refresh(ctx context.Context, claims *corev2.Claims) (*corev2.Claims, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Retrieve the right provider with the provider ID specified in the claims
	if provider, ok := a.providers[claims.Provider.ProviderID]; ok {
		user, err := provider.Refresh(ctx, claims)
		if err != nil {
			return nil, fmt.Errorf(
				"could not refresh user %q with provider %q: %s", claims.Provider.UserID, provider.Type(), err,
			)
		}

		return user, nil
	}

	return nil, fmt.Errorf(
		"could not refresh user %q with missing provider ID %q", claims.Provider.UserID, claims.Provider.ProviderID,
	)
}

// AddProvider adds a provided provider to the list of configured providers
func (a *Authenticator) AddProvider(provider corev2.AuthProvider) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Make sure the providers map is not nil
	if a.providers == nil {
		a.providers = map[string]corev2.AuthProvider{}
	}

	a.providers[provider.Name()] = provider
}

// Providers returns the configured providers
func (a *Authenticator) Providers() map[string]corev2.AuthProvider {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Create a new map and copy the authenticator providers into it in order to
	// prevent race conditions
	providers := make(map[string]corev2.AuthProvider, len(a.providers))
	for k, v := range a.providers {
		providers[k] = v
	}

	return providers
}

// RemoveProvider removes the provider with the given name from the list of
// configued providers
func (a *Authenticator) RemoveProvider(name string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if _, ok := a.providers[name]; !ok {
		return fmt.Errorf("provider %q is not configured, could not remove it", name)
	}

	// Make sure at least two providers are enabled so there's still one remaining
	if len(a.providers) == 1 {
		return fmt.Errorf("provider %q is the only provider configured, could not remove it ", name)
	}

	delete(a.providers, name)
	return nil
}

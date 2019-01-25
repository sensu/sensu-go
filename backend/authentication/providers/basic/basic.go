package basic

import (
	"context"
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/store"
)

// Type represents the type of the basic authentication provider
const Type = "basic"

// Provider represents the basic internal authentication provider
type Provider struct {
	Store store.Store
}

// Authenticate a user, with the provided credentials, against the Sensu store
func (p *Provider) Authenticate(ctx context.Context, username, password string) (*v2.Claims, error) {
	if username == "" || password == "" {
		return nil, errors.New("the username and the password must not be empty")
	}

	user, err := p.Store.AuthenticateUser(ctx, username, password)
	if err != nil {
		return nil, err
	}

	claims, err := jwt.NewClaims(user)
	if err != nil {
		return nil, err
	}

	// Set the provider claims
	claims.Provider = p.claims(user.Username)

	return claims, nil
}

// Refresh the claims of a user
func (p *Provider) Refresh(ctx context.Context, providerClaims v2.ProviderClaims) (*v2.Claims, error) {
	user, err := p.Store.GetUser(ctx, providerClaims.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("user %q does not exist", providerClaims.UserID)
	}

	claims, err := jwt.NewClaims(user)
	if err != nil {
		return nil, err
	}

	// Set the provider claims
	claims.Provider = p.claims(user.Username)

	return claims, nil
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "default"
}

// Type returns the provider type
func (p *Provider) Type() string {
	return Type
}

func (p *Provider) claims(username string) v2.ProviderClaims {
	return v2.ProviderClaims{
		ProviderID: v2.ProviderID(p),
		UserID:     username,
	}
}

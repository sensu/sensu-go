package basic

import (
	"context"
	"errors"
	"fmt"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/store"
)

// Type represents the type of the basic authentication provider
const Type = "basic"

// Provider represents the basic internal authentication provider
type Provider struct {
	Store store.Store

	// ObjectMeta contains the name, namespace, labels and annotations
	corev2.ObjectMeta `json:"metadata"`
}

// Authenticate a user, with the provided credentials, against the Sensu store
func (p *Provider) Authenticate(ctx context.Context, username, password string) (*corev2.Claims, error) {
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
func (p *Provider) Refresh(ctx context.Context, providerClaims corev2.AuthProviderClaims) (*corev2.Claims, error) {
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

// GetObjectMeta returns the provider metadata
func (p *Provider) GetObjectMeta() corev2.ObjectMeta {
	return p.ObjectMeta
}

// Name returns the provider name
func (p *Provider) Name() string {
	return p.ObjectMeta.Name
}

// Type returns the provider type
func (p *Provider) Type() string {
	return Type
}

// StorePrefix returns the path prefix to the provider in the store. Not
// implemented
func (p *Provider) StorePrefix() string {
	return ""
}

// URIPath returns the path component of the basic provider. Not implemented
func (p *Provider) URIPath() string {
	return ""
}

// Validate validates the basic provider configuration
func (p *Provider) Validate() error {
	p.ObjectMeta.Name = Type
	return nil
}

func (p *Provider) claims(username string) corev2.AuthProviderClaims {
	return corev2.AuthProviderClaims{
		ProviderID: p.Name(),
		UserID:     username,
	}
}

// SetNamespace sets the namespace of the resource.
func (p *Provider) SetNamespace(namespace string) {
	p.Namespace = namespace
}

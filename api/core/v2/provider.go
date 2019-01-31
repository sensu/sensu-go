package v2

import (
	"context"
	"fmt"
)

// AuthProvider represents an abstracted authentication provider
type AuthProvider interface {
	// Authenticate attempts to authenticate a user with its username and password
	Authenticate(ctx context.Context, username, password string) (*Claims, error)
	// Refresh renews the user claims with the provider claims
	Refresh(ctx context.Context, providerClaims AuthProviderClaims) (*Claims, error)

	// GetObjectMeta returns the object metadata for the provider
	GetObjectMeta() ObjectMeta
	// Name returns the provider name (e.g. default)
	Name() string
	// Type returns the provider type (e.g. ldap)
	Type() string
	// URIPath gives the path to the provider
	URIPath() string
	// Validate checks if the fields in the provider are valid
	Validate() error
}

// AuthProviderID returns a unique identifier for a given auth provider
func AuthProviderID(p AuthProvider) string {
	return fmt.Sprintf("%s/%s", p.Type(), p.Name())
}

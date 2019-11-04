package v2

import (
	"context"
)

// AuthProvider represents an abstracted authentication provider
type AuthProvider interface {
	Resource

	// Authenticate attempts to authenticate a user with its username and password
	Authenticate(ctx context.Context, username, password string) (*Claims, error)
	// Refresh renews the user claims with the provider claims
	Refresh(ctx context.Context, claims *Claims) (*Claims, error)

	// Name returns the provider name (e.g. default)
	Name() string
	// Type returns the provider type (e.g. ldap)
	Type() string
}

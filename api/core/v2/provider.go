package v2

import (
	"context"
	"fmt"
)

// AuthProvider represents an abstracted authentication provider
type AuthProvider interface {
	Authenticate(ctx context.Context, username, password string) (*Claims, error)
	Name() string
	Refresh(ctx context.Context, providerClaims AuthProviderClaims) (*Claims, error)
	Type() string
}

// AuthProviderID returns a unique identifier for a given auth provider
func AuthProviderID(p AuthProvider) string {
	return fmt.Sprintf("%s/%s", p.Type(), p.Name())
}

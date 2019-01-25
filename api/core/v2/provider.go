package v2

import (
	"context"
	fmt "fmt"
)

// Provider represents an abstracted authentication provider
type Provider interface {
	Authenticate(ctx context.Context, username, password string) (*Claims, error)
	Name() string
	Refresh(ctx context.Context, providerClaims ProviderClaims) (*Claims, error)
	Type() string
}

// ProviderID returns a unique identifier for a given provider
func ProviderID(p Provider) string {
	return fmt.Sprintf("%s/%s", p.Type(), p.Name())
}

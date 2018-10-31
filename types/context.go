package types

import "context"

// Define the key type to avoid key collisions in context
type key int

const (
	// AuthorizationActorKey is the key name used to store a user's details within
	// a context
	AuthorizationActorKey key = iota
	// AuthorizationRoleKey is the key name used to store a user's roles within
	// a context TODO: Remove
	AuthorizationRoleKey
	// AccessTokenClaims contains the key name to retrieve the access token claims
	AccessTokenClaims
	// ClaimsKey contains key name to retrieve the jwt claims from context
	ClaimsKey
	// NamespaceKey contains the key name to retrieve the namespace from context
	NamespaceKey
	// RefreshTokenClaims contains the key name to retrieve the refresh token claims
	RefreshTokenClaims
	// RefreshTokenString contains the key name to retrieve the refresh token string
	RefreshTokenString
	// StoreKey contains the key name to retrieve the etcd store from within a context
	StoreKey
)

// ContextNamespace returns the namespace injected in the context
func ContextNamespace(ctx context.Context) string {
	if value := ctx.Value(NamespaceKey); value != nil {
		return value.(string)
	}
	return ""
}

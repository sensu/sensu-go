package v2

import "context"

// Define the key type to avoid key collisions in context
type key int

const (
	// AccessTokenString is the key name used to retrieve the access token string
	AccessTokenString key = iota

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

	// AuthorizationAttributesKey is the key name used to store authorization
	// attributes extracted from a request
	AuthorizationAttributesKey

	// StoreKey contains the key name to retrieve the etcd store from within a context
	StoreKey

	// PageContinueKey contains the continue token used in pagination
	PageContinueKey

	// PageSizeKey contains the page size used in pagination
	PageSizeKey
)

// ContextNamespace returns the namespace injected in the context
func ContextNamespace(ctx context.Context) string {
	if value := ctx.Value(NamespaceKey); value != nil {
		return value.(string)
	}
	return ""
}

// PageSizeFromContext returns the page size stored in the given context, if
// any. Returns 0 if none is found, typically meaning "unlimited" page size.
func PageSizeFromContext(ctx context.Context) int {
	if value := ctx.Value(PageSizeKey); value != nil {
		return value.(int)
	}
	return 0
}

// PageContinueFromContext returns the continue token stored in the given
// context, if any. Returns "" if none is found.
func PageContinueFromContext(ctx context.Context) string {
	if value := ctx.Value(PageContinueKey); value != nil {
		return value.(string)
	}
	return ""
}

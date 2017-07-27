package types

// Define the key type to avoid key collisions in context
type key int

const (
	// OrganizationKey contains the key name to retrieve the org from context
	OrganizationKey key = iota
	// AuthorizationRoleKey is the key name used to store a user's roles within
	// a context
	AuthorizationRoleKey
	// AccessTokenClaims contains the key name to retrieve the access token claims
	AccessTokenClaims
	// RefreshTokenClaims contains the key name to retrieve the refresh token claims
	RefreshTokenClaims
	// RefreshTokenString contains the key name to retrieve the refresh token string
	RefreshTokenString
)

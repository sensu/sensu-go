package types

// Define the key type to avoid key collisions in context
type key int

const (
	// OrganizationKey contains the key name to retrieve the org from context
	OrganizationKey key = iota
	// AuthorizationRoleKey is the key name used to store a user's roles within
	// a context TODO: Remove
	AuthorizationRoleKey
	// AuthorizationActorKey is the key name used to store a user's details within
	// a context
	AuthorizationActorKey
	// AccessTokenClaims contains the key name to retrieve the access token claims
	AccessTokenClaims
	// RefreshTokenClaims contains the key name to retrieve the refresh token claims
	RefreshTokenClaims
	// RefreshTokenString contains the key name to retrieve the refresh token string
	RefreshTokenString
	// ClaimsKey contains key name to retrieve the jwt claims from context
	ClaimsKey
)

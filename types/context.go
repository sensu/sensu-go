package types

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
	// EnvironmentKey contains the key name to retrieve the env from context
	EnvironmentKey
	// OrganizationKey contains the key name to retrieve the org from context
	OrganizationKey
	// RefreshTokenClaims contains the key name to retrieve the refresh token claims
	RefreshTokenClaims
	// RefreshTokenString contains the key name to retrieve the refresh token string
	RefreshTokenString
	// StoreKey contains the key name to retrieve the etcd store from within a context
	StoreKey
)

package v2

import (
	"errors"

	jwt "github.com/golang-jwt/jwt/v4"
)

var (
	ErrInvalidToken = errors.New("invalid access or refresh token")
	ErrUnauthorized = errors.New("unauthorized")
)

// Claims represents the JWT claims
type Claims struct {
	jwt.StandardClaims

	// Custom claims
	Groups   []string           `json:"groups"`
	Provider AuthProviderClaims `json:"provider"`
	APIKey   bool               `json:"api_key"`
}

// AuthProviderClaims contains information from the authentication provider
type AuthProviderClaims struct {
	// ProviderID used to login the user
	ProviderID string `json:"provider_id"`
	// ProviderType represents the type of provider used
	ProviderType string `json:"provider_type"`
	// UserID assigned to the user by the provider
	UserID string `json:"user_id"`
}

// FixtureClaims returns a testing fixture for a JWT claims
func FixtureClaims(username string, groups []string) *Claims {
	return &Claims{
		StandardClaims: jwt.StandardClaims{Subject: username},
		Groups:         groups,
		Provider: AuthProviderClaims{
			ProviderID:   "basic/default",
			ProviderType: "basic",
			UserID:       username,
		},
	}
}

// StandardClaims returns an initialized jwt.StandardClaims with the subject
func StandardClaims(subject string) jwt.StandardClaims {
	return jwt.StandardClaims{Subject: subject}
}

package v2

import jwt "github.com/dgrijalva/jwt-go"

// Claims represents the JWT claims
type Claims struct {
	jwt.StandardClaims

	// Custom claims
	Groups   []string       `json:"groups"`
	Provider ProviderClaims `json:"provider"`
}

// ProviderClaims contains information from the authentication provider
type ProviderClaims struct {
	// ProviderID used to login the user
	ProviderID string `json:"provider_id"`
	// UserID assigned to the user by the provider
	UserID string `json:"user_id"`
}

// FixtureClaims returns a testing fixture for a JWT claims
func FixtureClaims(username string, groups []string) *Claims {
	return &Claims{
		StandardClaims: jwt.StandardClaims{Subject: username},
		Groups:         groups,
		Provider: ProviderClaims{
			ProviderID: "basic/default",
			UserID:     username,
		},
	}
}

// StandardClaims returns an initialized jwt.StandardClaims with the subject
func StandardClaims(subject string) jwt.StandardClaims {
	return jwt.StandardClaims{Subject: subject}
}

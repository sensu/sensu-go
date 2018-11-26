package v2

import jwt "github.com/dgrijalva/jwt-go"

// Claims represents the JWT claims
type Claims struct {
	jwt.StandardClaims
	Groups []string
}

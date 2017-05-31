package types

import jwt "github.com/dgrijalva/jwt-go"

// Claims represents the JWT claims
type Claims struct {
	jwt.StandardClaims
}

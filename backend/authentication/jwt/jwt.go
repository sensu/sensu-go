package jwt

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/context"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	utilbytes "github.com/sensu/sensu-go/util/bytes"
)

var secret []byte

const (
	// claimsKey contains the key name used to store the JWT claims within
	// the context of a request
	claimsKey = "JWTClaims"
)

// AccessToken creates a new access token and returns it in both JWT and
// signed format, along with any error
func AccessToken(username string) (*jwt.Token, string, error) {
	// Create a unique identifier for the token
	jti, err := utilbytes.Random(16)
	if err != nil {
		return nil, "", err
	}

	claims := types.Claims{
		StandardClaims: jwt.StandardClaims{
			Id:       hex.EncodeToString(jti),
			IssuedAt: time.Date(2015, 10, 10, 12, 0, 0, 0, time.UTC).Unix(),
			Subject:  username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token as a string using the secret
	tokenString, err := token.SignedString(secret)
	if err != nil {
		return nil, "", err
	}

	return token, tokenString, nil
}

// GetClaims returns the claims from a token
func GetClaims(token *jwt.Token) (*types.Claims, error) {
	if claims, ok := token.Claims.(types.Claims); ok {
		return &claims, nil
	}

	return nil, fmt.Errorf("Could not parse the token claims")
}

// GetClaimsFromContext retrieves the JWT claims from the request context
func GetClaimsFromContext(r *http.Request) *types.Claims {
	if value := context.Get(r, claimsKey); value != nil {
		claims, ok := value.(types.Claims)
		if !ok {
			return nil
		}
		return &claims
	}
	return nil
}

// InitSecret initializes and retrieves the secret for our signing tokens
func InitSecret(store store.Store) error {
	var s []byte
	var err error

	// Retrieve the secret
	if secret == nil {
		s, err = store.GetJWTSecret()
		if err != nil {
			// The secret does not exist, we need to create one
			s, err = utilbytes.Random(32)
			if err != nil {
				return err
			}

			// Add the secret to the store
			err = store.CreateJWTSecret(s)
			if err != nil {
				return err
			}
		}

		// Set the secret so it's available accross the package
		secret = s
	}

	return nil
}

// ParseToken takes a signed token and parse it to verify its integrity
func ParseToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(tokenString, &types.Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// secret is a []byte containing the secret
		return secret, nil
	})
	if err != nil {
		return nil, err
	}

	if _, ok := token.Claims.(*types.Claims); !ok || !token.Valid {
		return nil, fmt.Errorf("Invalid JSON Web Token")
	}

	return token, nil
}

// SetClaimsIntoContext adds the token claims into the request context for
// easier consumption later
func SetClaimsIntoContext(r *http.Request, token *jwt.Token) {
	claims, _ := token.Claims.(types.Claims)
	context.Set(r, claimsKey, claims)
}

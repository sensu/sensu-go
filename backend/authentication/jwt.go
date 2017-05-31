package authentication

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
	claimsKey = "JWTClaims"
	issuer    = "sensu.io"
)

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

// getClaimsFromContext retrieves the JWT claims from the request context
func getClaimsFromContext(r *http.Request) jwt.MapClaims {
	if value := context.Get(r, claimsKey); value != nil {
		return value.(jwt.MapClaims)
	}
	return nil
}

// newToken creates a new signed token
func newToken(user *types.User) (*jwt.Token, string, error) {
	// Create a unique identifier for the token
	jti, err := utilbytes.Random(16)
	if err != nil {
		return nil, "", err
	}

	// Create a new token object, specifying signing method and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iat": time.Date(2015, 10, 10, 12, 0, 0, 0, time.UTC).Unix(),
		"iss": issuer,
		"jti": hex.EncodeToString(jti),
		"sub": user.Username,
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString(secret)
	if err != nil {
		return nil, "", err
	}

	return token, tokenString, nil
}

// parseToken takes the token string and parse it to verify its integrity
func parseToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
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

	if _, ok := token.Claims.(jwt.MapClaims); !ok || !token.Valid {
		return nil, fmt.Errorf("Invalid JSON Web Token")
	}

	return token, nil
}

// setClaimsIntoContext adds the token claims into the request context for
// easier consumption later
func setClaimsIntoContext(r *http.Request, token *jwt.Token) {
	claims, _ := token.Claims.(jwt.MapClaims)
	context.Set(r, claimsKey, claims)
}

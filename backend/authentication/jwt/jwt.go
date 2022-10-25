package jwt

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	time "github.com/echlebek/timeproxy"
	jwt "github.com/golang-jwt/jwt/v4"
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/types"
	utilbytes "github.com/sensu/sensu-go/util/bytes"
)

type key int

const (
	// IssuerURLKey specifies the URL on which the JWT is issued
	IssuerURLKey key = iota
)

const Name = "jwt"

var (
	defaultExpiration = time.Minute * 5
	secret            []byte
	privateKey        *ecdsa.PrivateKey
	publicKey         *ecdsa.PublicKey
	signingMethod     jwt.SigningMethod
)

func init() {
	signingMethod = jwt.SigningMethodHS256

	var err error

	// secret should be initialized and persisted, but in case that fails,
	// it's critical to have a non-empty jwt secret
	secret, err = utilbytes.Random(32)
	if err != nil {
		panic(err)
	}
}

// AccessToken creates a new access token and returns it in both JWT and
// signed format, along with any error
func AccessToken(claims *corev2.Claims) (*jwt.Token, string, error) {
	// Create a unique identifier for the token
	jti, err := GenJTI()
	if err != nil {
		return nil, "", err
	}
	claims.Id = jti

	// Add an expiration to the token
	claims.ExpiresAt = time.Now().Add(defaultExpiration).Unix()

	token := jwt.NewWithClaims(signingMethod, claims)

	// Determine which key to use to sign the token
	var key interface{}
	if signingMethod == jwt.SigningMethodHS256 {
		key = secret
	} else {
		key = privateKey
	}

	// Sign the token
	tokenString, err := token.SignedString(key)
	if err != nil {
		return nil, "", err
	}
	return token, tokenString, nil
}

// NewClaims creates new claim based on username
func NewClaims(user *corev2.User) (*corev2.Claims, error) {
	// Create a unique identifier for the token
	jti, err := GenJTI()
	if err != nil {
		return nil, err
	}

	claims := &corev2.Claims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(defaultExpiration).Unix(),
			Id:        jti,
			Subject:   user.Username,
		},
		Groups: user.Groups,
	}
	return claims, nil
}

// GenJTI generates a new random JTI
func GenJTI() (string, error) {
	jti, err := utilbytes.Random(16)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(jti), err
}

// GetClaims returns the claims from a token
func GetClaims(token *jwt.Token) (*corev2.Claims, error) {
	if claims, ok := token.Claims.(*corev2.Claims); ok {
		return claims, nil
	}

	return nil, fmt.Errorf("could not parse the token claims")
}

// GetClaimsFromContext retrieves the JWT claims from the request context
func GetClaimsFromContext(ctx context.Context) *corev2.Claims {
	if value := ctx.Value(corev2.ClaimsKey); value != nil {
		claims, ok := value.(*corev2.Claims)
		if !ok {
			return nil
		}
		return claims
	}
	return nil
}

// ExtractBearerToken retrieves the bearer token from a request and returns the
// JWT
func ExtractBearerToken(r *http.Request) string {
	// Does a bearer token was provided in the Authorization header?
	var tokenString string
	tokens, ok := r.Header["Authorization"]
	if ok && len(tokens) >= 1 {
		tokenString = tokens[0]
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	}

	return tokenString
}

// LoadKeyPair loads a private and public key pair from files
func LoadKeyPair(privatePath, publicPath string) error {
	if privatePath != "" && publicPath == "" {
		return errors.New("a public key is required when specifying a private key")
	}

	if publicPath != "" {
		publicBytes, err := ioutil.ReadFile(publicPath)
		if err != nil {
			return fmt.Errorf("unable to read the public key file: %s", err)
		}
		if publicKey, err = jwt.ParseECPublicKeyFromPEM(publicBytes); err != nil {
			return fmt.Errorf("unable to parse the ECDSA public key: %v", err)
		}
	}

	if privatePath != "" {
		privateBytes, err := ioutil.ReadFile(privatePath)
		if err != nil {
			return fmt.Errorf("unable to read the private key file: %s", err)
		}
		if privateKey, err = jwt.ParseECPrivateKeyFromPEM(privateBytes); err != nil {
			return fmt.Errorf("unable to parse the ECDSA private key: %v", err)
		}

		// Determine the signing method to use
		switch bitSize := publicKey.Curve.Params().BitSize; bitSize {
		case 256:
			signingMethod = jwt.SigningMethodES256
		case 384:
			signingMethod = jwt.SigningMethodES384
		case 521:
			signingMethod = jwt.SigningMethodES512
		default:
			return fmt.Errorf("could not determine a signing method for curve %s", publicKey.Curve.Params().Name)
		}
	}

	return nil
}

// SetSecret does something incredibly dirty - sets a global variable.
func SetSecret(s []byte) {
	secret = s
}

// parseToken takes a signed token and parse it to verify its integrity
func parseToken(tokenString string) (*jwt.Token, error) {
	t, err := jwt.ParseWithClaims(tokenString, &types.Claims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Header["alg"] == jwt.SigningMethodHS256.Alg() {
			// Validate the signing method used
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			// Use the secret to verify the signature
			return secret, nil
		}

		// Validate that we do have a public key available
		if publicKey == nil {
			return nil, errors.New("no public key available to validate the signature")
		}

		// Validate the signing method used
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Validate the algorith used
		switch alg := token.Header["alg"]; alg {
		case jwt.SigningMethodES256.Alg(), jwt.SigningMethodES384.Alg(), jwt.SigningMethodES512.Alg():
		default:
			return nil, fmt.Errorf("unexpected algorith %q found in token header", alg)
		}

		// Use the public key to verify the signature
		return publicKey, nil
	})
	return t, err
}

// RefreshToken returns a refresh token for a specific user
func RefreshToken(claims *corev2.Claims) (*jwt.Token, string, error) {
	// Create a unique identifier for the token
	jti, err := GenJTI()
	if err != nil {
		return nil, "", err
	}
	claims.Id = jti

	token := jwt.NewWithClaims(signingMethod, claims)

	// Determine which key to use to sign the token
	var key interface{}
	if signingMethod == jwt.SigningMethodHS256 {
		key = secret
	} else {
		key = privateKey
	}

	// Sign the token
	tokenString, err := token.SignedString(key)
	if err != nil {
		return nil, "", err
	}

	return token, tokenString, nil
}

// SetClaimsIntoContext adds the token claims into the request context for
// easier consumption later
func SetClaimsIntoContext(r *http.Request, claims *corev2.Claims) context.Context {
	return context.WithValue(r.Context(), corev2.ClaimsKey, claims)
}

// ValidateExpiredToken verifies that the provided token is valid, even if
// it's expired
func ValidateExpiredToken(tokenString string) (*jwt.Token, error) {
	token, err := parseToken(tokenString)
	if token == nil {
		return nil, err
	}

	if _, ok := token.Claims.(*corev2.Claims); ok {
		if token.Valid {
			return token, nil
		}

		// Inspect the error to determine the cause
		if validationError, ok := err.(*jwt.ValidationError); ok {
			if validationError.Errors&jwt.ValidationErrorExpired != 0 {
				// We already know that the token is expired and we don't care at that
				// point, we simply want to know if there's any other error
				validationError.Errors ^= jwt.ValidationErrorExpired
			}

			// Return the token if we have no other validation error
			if validationError.Errors == 0 {
				return token, nil
			}
		}
	}

	return nil, err
}

// ValidateToken verifies that the provided token is valid
func ValidateToken(tokenString string) (*jwt.Token, error) {
	token, err := parseToken(tokenString)
	if token == nil {
		return nil, err
	}
	if _, ok := token.Claims.(*types.Claims); ok && token.Valid {
		return token, nil
	}
	return nil, err
}

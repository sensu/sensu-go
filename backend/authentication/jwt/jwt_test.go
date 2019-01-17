package jwt

import (
	"fmt"
	"net/http"
	"testing"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/echlebek/crock"
	time "github.com/echlebek/timeproxy"
	"github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

var testTime = crock.NewTime(time.Now())

func init() {
	time.TimeProxy = testTime
	jwt.TimeFunc = testTime.Now
}

func TestAccessToken(t *testing.T) {
	secret = []byte("foobar")
	claims := &v2.Claims{StandardClaims: jwt.StandardClaims{Subject: "foo"}}

	_, tokenString, err := AccessToken(claims)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	token, err := ValidateToken(tokenString)
	assert.NoError(t, err)
	assert.NotNil(t, token)

	tokenClaims, _ := token.Claims.(*types.Claims)
	assert.Equal(t, claims.Subject, tokenClaims.Subject)
	assert.NotEmpty(t, tokenClaims.Id)
	assert.NotZero(t, tokenClaims.ExpiresAt)
}

func TestClaimsContext(t *testing.T) {
	claims := &v2.Claims{StandardClaims: jwt.StandardClaims{Subject: "foo"}}

	r, _ := http.NewRequest("GET", "/foo", nil)
	ctx := SetClaimsIntoContext(r, claims)

	tokenClaims := GetClaimsFromContext(ctx)
	assert.Equal(t, claims.Subject, tokenClaims.Subject)
}

func TestGetClaims(t *testing.T) {
	claims := &v2.Claims{StandardClaims: jwt.StandardClaims{Subject: "foo"}}
	token, _, _ := AccessToken(claims)

	_, err := GetClaims(token)
	assert.NoError(t, err)
}

func TestExtractBearerToken(t *testing.T) {
	claims := &v2.Claims{StandardClaims: jwt.StandardClaims{Subject: "foo"}}

	// No bearer token
	r, _ := http.NewRequest("GET", "/foo", nil)
	token := ExtractBearerToken(r)

	assert.Empty(t, token)

	// Valid bearer token
	r, _ = http.NewRequest("GET", "/foo", nil)
	_, tokenString, _ := AccessToken(claims)
	r.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokenString))
	token = ExtractBearerToken(r)

	assert.NotEmpty(t, token)
}

func TestInitSecret(t *testing.T) {
	secret = nil
	store := &mockstore.MockStore{}
	store.On("GetJWTSecret").Return("foo", nil)

	err := InitSecret(store)
	assert.NoError(t, err)
	assert.NotEqual(t, nil, secret)
}

func TestInitSecretMissingSecret(t *testing.T) {
	secret = nil
	store := &mockstore.MockStore{}
	store.On("GetJWTSecret").Return("", fmt.Errorf(""))
	store.On("CreateJWTSecret").Return(nil)

	err := InitSecret(store)
	assert.NoError(t, err)
	assert.NotEqual(t, nil, secret)
}

func TestInitSecretEtcdError(t *testing.T) {
	secret = nil
	store := &mockstore.MockStore{}
	store.On("GetJWTSecret").Return("", fmt.Errorf(""))
	store.On("CreateJWTSecret").Return(fmt.Errorf(""))

	err := InitSecret(store)
	assert.Error(t, err)
	assert.Equal(t, []byte(nil), secret)
}

func TestRefreshToken(t *testing.T) {
	secret = []byte("foobar")
	claims := &v2.Claims{StandardClaims: jwt.StandardClaims{Subject: "foo"}}
	_, tokenString, err := RefreshToken(claims)
	assert.NoError(t, err)

	token, err := ValidateToken(tokenString)
	assert.NoError(t, err)
	assert.NotNil(t, token)

	tokenClaims, _ := token.Claims.(*types.Claims)
	assert.Equal(t, claims.Subject, tokenClaims.Subject)
	assert.NotEmpty(t, tokenClaims.Id)
}

func TestValidateTokenError(t *testing.T) {
	// Create an expired token
	claims := &v2.Claims{StandardClaims: jwt.StandardClaims{Subject: "foo"}}
	_, tokenString, _ := AccessToken(claims)

	// Assert that the token is currently valid
	_, err := ValidateToken(tokenString)
	assert.NoError(t, err)

	// The token should expire after the expiration time
	testTime.Set(time.Now().Add(defaultExpiration + time.Hour))
	_, err = ValidateToken(tokenString)
	assert.Error(t, err)
}

// ValidateExpiredToken should not return an error if we provide an expired
// token but otherwise valid
func TestValidateExpiredToken(t *testing.T) {
	// Create an expired token
	claims := &v2.Claims{StandardClaims: jwt.StandardClaims{Subject: "foo"}}
	_, tokenString, _ := AccessToken(claims)

	// Wait for the token to expire
	testTime.Set(time.Now().Add(defaultExpiration + time.Second))
	_, err := ValidateExpiredToken(tokenString)
	assert.NoError(t, err, "An expired token should not be considered as invalid")
}

func TestValidateExpiredTokenActive(t *testing.T) {
	claims := &v2.Claims{StandardClaims: jwt.StandardClaims{Subject: "foo"}}

	_, tokenString, err := AccessToken(claims)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	token, err := ValidateToken(tokenString)
	assert.NoError(t, err, "A valid token should not be considered as invalid")
	assert.NotNil(t, token)
}

func TestValidateExpiredTokenInvalid(t *testing.T) {
	// Create an expired token
	secret = []byte("foobar")
	claims := &v2.Claims{StandardClaims: jwt.StandardClaims{Subject: "foo"}}
	_, tokenString, _ := AccessToken(claims)

	// The token will expire
	testTime.Set(time.Now().Add(defaultExpiration + time.Second))

	// Modify the secret so it's no longer valid
	secret = []byte("qux")

	_, err := ValidateExpiredToken(tokenString)
	assert.Error(t, err, "An invalid token should not be valid even if it's expired")
}

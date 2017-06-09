package jwt

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/assert"
)

func TestAccessToken(t *testing.T) {
	secret = []byte("foobar")
	username := "foo"

	_, tokenString, err := AccessToken(username)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	token, err := ValidateToken(tokenString)
	assert.NoError(t, err)
	assert.NotNil(t, token)

	claims, _ := token.Claims.(*Claims)
	assert.Equal(t, username, claims.Subject)
	assert.NotEmpty(t, claims.Id)
	assert.NotZero(t, claims.ExpiresAt)
}

func TestClaimsContext(t *testing.T) {
	username := "foo"
	token, _, _ := AccessToken(username)

	r, _ := http.NewRequest("GET", "/foo", nil)

	SetClaimsIntoContext(r, token)
	claims := GetClaimsFromContext(r)
	assert.Equal(t, username, claims.Subject)
}

func TestGetClaims(t *testing.T) {
	username := "foo"
	token, _, _ := AccessToken(username)

	_, err := GetClaims(token)
	assert.NoError(t, err)
}

func TestExtractBearerToken(t *testing.T) {
	// No bearer token
	r, _ := http.NewRequest("GET", "/foo", nil)
	token := ExtractBearerToken(r)

	assert.Empty(t, token)

	// Valid bearer token
	r, _ = http.NewRequest("GET", "/foo", nil)
	_, tokenString, _ := AccessToken("foo")
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
	username := "foo"
	tokenString, err := RefreshToken(username)
	assert.NoError(t, err)

	token, err := ValidateToken(tokenString)
	assert.NoError(t, err)
	assert.NotNil(t, token)

	claims, _ := token.Claims.(*Claims)
	assert.Equal(t, username, claims.Subject)
	assert.NotEmpty(t, claims.Id)
}

func TestValidateTokenError(t *testing.T) {
	// Create an expired token
	defaultExpiration = time.Second * time.Duration(1)
	username := "foo"
	_, tokenString, _ := AccessToken(username)

	// Wait for the token to expire
	time.Sleep(time.Second * 2)
	_, err := ValidateToken(tokenString)
	assert.Error(t, err)

	// Set back the default value
	defaultExpiration = time.Minute * time.Duration(15)
}

// ValidateExpiredToken should not return an error if we provide an expired
// token but otherwise valid
func TestValidateExpiredToken(t *testing.T) {
	// Create an expired token
	defaultExpiration = time.Second * time.Duration(1)
	username := "foo"
	_, tokenString, _ := AccessToken(username)

	// Wait for the token to expire
	time.Sleep(time.Second * 2)
	_, err := ValidateExpiredToken(tokenString)
	assert.NoError(t, err, "An expired token should not be considered as invalid")

	// Set back the default value
	defaultExpiration = time.Minute * time.Duration(15)
}

func TestValidateExpiredTokenActive(t *testing.T) {
	username := "foo"

	_, tokenString, err := AccessToken(username)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	token, err := ValidateToken(tokenString)
	assert.NoError(t, err, "A valid token should not be considered as invalid")
	assert.NotNil(t, token)
}

func TestValidateExpiredTokenInvalid(t *testing.T) {
	// Create an expired token
	secret = []byte("foobar")
	defaultExpiration = time.Second * time.Duration(1)
	username := "foo"
	_, tokenString, _ := AccessToken(username)

	// Wait for the token to expire
	time.Sleep(time.Second * 2)

	// Modify the secret so it's no longer valid
	secret = []byte("qux")

	_, err := ValidateExpiredToken(tokenString)
	assert.Error(t, err, "An invalid token should not be valid even if it's expired")

	// Set back the default value
	defaultExpiration = time.Minute * time.Duration(15)
}

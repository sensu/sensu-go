package jwt

import (
	"fmt"
	"net/http"
	"testing"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/echlebek/crock"
	time "github.com/echlebek/timeproxy"
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
	user := &types.User{Username: "foo"}

	_, tokenString, err := AccessToken(user)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	token, err := ValidateToken(tokenString)
	assert.NoError(t, err)
	assert.NotNil(t, token)

	claims, _ := token.Claims.(*types.Claims)
	assert.Equal(t, user.Username, claims.Subject)
	assert.NotEmpty(t, claims.Id)
	assert.NotZero(t, claims.ExpiresAt)
}

func TestClaimsContext(t *testing.T) {
	user := &types.User{Username: "foo"}
	token, _, _ := AccessToken(user)

	r, _ := http.NewRequest("GET", "/foo", nil)

	ctx := SetClaimsIntoContext(r, token.Claims.(*types.Claims))
	claims := GetClaimsFromContext(ctx)
	assert.Equal(t, user.Username, claims.Subject)
}

func TestGetClaims(t *testing.T) {
	user := &types.User{Username: "foo"}
	token, _, _ := AccessToken(user)

	_, err := GetClaims(token)
	assert.NoError(t, err)
}

func TestExtractBearerToken(t *testing.T) {
	user := &types.User{Username: "foo"}

	// No bearer token
	r, _ := http.NewRequest("GET", "/foo", nil)
	token := ExtractBearerToken(r)

	assert.Empty(t, token)

	// Valid bearer token
	r, _ = http.NewRequest("GET", "/foo", nil)
	_, tokenString, _ := AccessToken(user)
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
	user := &types.User{Username: "foo"}
	_, tokenString, err := RefreshToken(user)
	assert.NoError(t, err)

	token, err := ValidateToken(tokenString)
	assert.NoError(t, err)
	assert.NotNil(t, token)

	claims, _ := token.Claims.(*types.Claims)
	assert.Equal(t, user.Username, claims.Subject)
	assert.NotEmpty(t, claims.Id)
}

func TestValidateTokenError(t *testing.T) {
	// Create an expired token
	user := &types.User{Username: "foo"}
	_, tokenString, _ := AccessToken(user)

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
	user := &types.User{Username: "foo"}
	_, tokenString, _ := AccessToken(user)

	// Wait for the token to expire
	testTime.Set(time.Now().Add(defaultExpiration + time.Second))
	_, err := ValidateExpiredToken(tokenString)
	assert.NoError(t, err, "An expired token should not be considered as invalid")
}

func TestValidateExpiredTokenActive(t *testing.T) {
	user := &types.User{Username: "foo"}

	_, tokenString, err := AccessToken(user)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	token, err := ValidateToken(tokenString)
	assert.NoError(t, err, "A valid token should not be considered as invalid")
	assert.NotNil(t, token)
}

func TestValidateExpiredTokenInvalid(t *testing.T) {
	// Create an expired token
	secret = []byte("foobar")
	user := &types.User{Username: "foo"}
	_, tokenString, _ := AccessToken(user)

	// The token will expire
	testTime.Set(time.Now().Add(defaultExpiration + time.Second))

	// Modify the secret so it's no longer valid
	secret = []byte("qux")

	_, err := ValidateExpiredToken(tokenString)
	assert.Error(t, err, "An invalid token should not be valid even if it's expired")
}

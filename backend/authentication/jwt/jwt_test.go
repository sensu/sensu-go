package jwt

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestAccessTokenAndParseToken(t *testing.T) {
	secret = []byte("foobar")
	username := "foo"

	_, tokenString, err := AccessToken(username)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	token, err := ParseToken(tokenString)
	assert.NoError(t, err)
	assert.NotNil(t, token)

	claims, _ := token.Claims.(*types.Claims)
	assert.Equal(t, username, claims.Subject)
	assert.Equal(t, issuer, claims.Issuer)
	assert.NotEmpty(t, claims.Id)
	assert.NotZero(t, claims.IssuedAt)
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

	claims, err := GetClaims(token)
	assert.NoError(t, err)
	assert.Equal(t, username, claims.Subject)
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

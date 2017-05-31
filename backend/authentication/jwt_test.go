package authentication

import (
	"fmt"
	"net/http"
	"testing"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestClaimsContext(t *testing.T) {
	user := types.FixtureUser("foo")
	token, _, _ := newToken(user)

	r, _ := http.NewRequest("GET", "/foo", nil)

	setClaimsIntoContext(r, token)
	claims := getClaimsFromContext(r)
	assert.Equal(t, user.Username, claims["sub"])
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

func TestNewTokenAndParseToken(t *testing.T) {
	var date float64

	secret = []byte("foobar")
	user := types.FixtureUser("foo")

	_, tokenString, err := newToken(user)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	token, err := parseToken(tokenString)
	assert.NoError(t, err)
	assert.NotNil(t, token)

	claims, _ := token.Claims.(jwt.MapClaims)
	assert.Equal(t, user.Username, claims["sub"])
	assert.Equal(t, issuer, claims["iss"])
	assert.NotEmpty(t, claims["jti"])
	assert.IsType(t, date, claims["iat"])
}

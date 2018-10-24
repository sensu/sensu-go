package middlewares

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestAllowList(t *testing.T) {
	// Create a token
	user := &types.User{Username: "foo"}
	token, tokenString, _ := jwt.AccessToken(user)
	claims, _ := jwt.GetClaims(token)

	store := &mockstore.MockStore{}
	store.On("GetToken", claims.Subject, claims.Id).Return(claims, nil)

	authMware := Authentication{}
	allowMware := AllowList{Store: store}
	server := httptest.NewServer(authMware.Then(allowMware.Then(testHandler())))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL, nil)
	// Add the bearer token in the Authorization header
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokenString))

	// Perform the request with the middleware
	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestMissingTokenFromAllowList(t *testing.T) {
	// Create a token
	user := &types.User{Username: "foo"}
	token, tokenString, _ := jwt.AccessToken(user)
	claims, _ := jwt.GetClaims(token)

	store := &mockstore.MockStore{}
	store.On("GetToken", claims.Subject, claims.Id).Return(claims, fmt.Errorf("error"))

	auth := Authentication{}
	allow := AllowList{Store: store}
	server := httptest.NewServer(auth.Then(allow.Then(testHandler())))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL, nil)
	// Add the bearer token in the Authorization header
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokenString))

	// Perform the request with the middleware
	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
}

func TestAllowListNoTokenIntoContext(t *testing.T) {
	store := &mockstore.MockStore{}

	allow := AllowList{Store: store}
	server := httptest.NewServer(allow.Then(testHandler()))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL, nil)

	// Perform the request with the middleware
	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
}

package middlewares

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/assert"
)

func TestAllowList(t *testing.T) {
	// Create a token
	claims := corev2.FixtureClaims("foo", nil)
	_, tokenString, _ := jwt.AccessToken(claims)

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
	claims := corev2.FixtureClaims("bar", nil)
	_, tokenString, _ := jwt.AccessToken(claims)

	sErr := &store.ErrNotFound{}
	store := &mockstore.MockStore{}
	store.On("GetToken", claims.Subject, claims.Id).Return(claims, sErr)

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

func TestAllowListIgnoreMissingClaims(t *testing.T) {
	store := &mockstore.MockStore{}

	allow := AllowList{Store: store, IgnoreMissingClaims: true}
	server := httptest.NewServer(allow.Then(testHandler()))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL, nil)

	// Perform the request with the middleware
	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

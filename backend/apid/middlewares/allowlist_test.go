package middlewares

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/assert"
)

func TestAllowList(t *testing.T) {
	// Create a token
	token, tokenString, _ := jwt.AccessToken("foo")
	claims, _ := jwt.GetClaims(token)

	store := &mockstore.MockStore{}
	store.On("GetToken", claims.Subject, claims.Id).Return(claims, nil)

	authMware := Authentication{}
	server := httptest.NewServer(authMware.Register(AllowList(testHandler(), store)))
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
	token, tokenString, _ := jwt.AccessToken("foo")
	claims, _ := jwt.GetClaims(token)

	store := &mockstore.MockStore{}
	store.On("GetToken", claims.Subject, claims.Id).Return(claims, fmt.Errorf("error"))

	authMware := Authentication{}
	server := httptest.NewServer(authMware.Register(AllowList(testHandler(), store)))
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

	server := httptest.NewServer(AllowList(testHandler(), store))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL, nil)

	// Perform the request with the middleware
	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
}

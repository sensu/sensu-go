package middlewares

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/stretchr/testify/assert"
)

func TestMiddlewareNoCredentials(t *testing.T) {
	mware := Authentication{}
	server := httptest.NewServer(mware.Then(testHandler()))
	defer server.Close()

	// No credentials passed
	res, err := http.Get(server.URL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
}

func TestMiddlewareJWT(t *testing.T) {
	mware := Authentication{}
	server := httptest.NewServer(mware.Then(testHandler()))
	defer server.Close()

	// Valid JWT
	claims := v2.FixtureClaims("foo", nil)
	_, tokenString, _ := jwt.AccessToken(claims)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", server.URL, nil)

	// Add the bearer token in the Authorization header
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokenString))

	res, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestMiddlewareInvalidJWT(t *testing.T) {
	mware := Authentication{}
	server := httptest.NewServer(mware.Then(testHandler()))
	defer server.Close()

	// Valid JWT
	tokenString := "foobar"

	client := &http.Client{}
	req, _ := http.NewRequest("GET", server.URL, nil)

	// Add the bearer token in the Authorization header
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokenString))

	res, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
}

func TestMiddlewareIgnoreUnauthorized(t *testing.T) {
	mware := Authentication{IgnoreUnauthorized: true}
	server := httptest.NewServer(mware.Then(testHandler()))
	defer server.Close()

	// No credentials passed
	res, err := http.Get(server.URL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

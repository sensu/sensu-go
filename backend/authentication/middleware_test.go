package authentication

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/testing/mockprovider"
	"github.com/stretchr/testify/assert"
)

func testHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Success")
		return
	})
}

func TestMiddlewareDisabledAuth(t *testing.T) {
	provider := &mockprovider.MockProvider{}
	server := httptest.NewServer(Middleware(provider, testHandler()))
	defer server.Close()

	// Disabled authentication
	provider.On("AuthEnabled").Return(false)
	res, err := http.Get(server.URL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestMiddlewareNoCredentials(t *testing.T) {
	provider := &mockprovider.MockProvider{}
	server := httptest.NewServer(Middleware(provider, testHandler()))
	defer server.Close()

	// No credentials passed
	provider.On("AuthEnabled").Return(true)
	res, err := http.Get(server.URL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
}

func TestMiddlewareJWT(t *testing.T) {
	provider := &mockprovider.MockProvider{}
	server := httptest.NewServer(Middleware(provider, testHandler()))
	defer server.Close()

	provider.On("AuthEnabled").Return(true)

	// Valid JWT
	_, tokenString, _ := jwt.AccessToken("foo")

	client := &http.Client{}
	req, _ := http.NewRequest("GET", server.URL, nil)

	// Add the bearer token in the Authorization header
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokenString))

	res, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestMiddlewareInvalidJWT(t *testing.T) {
	provider := &mockprovider.MockProvider{}
	server := httptest.NewServer(Middleware(provider, testHandler()))
	defer server.Close()

	provider.On("AuthEnabled").Return(true)

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

package authentication

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sensu/sensu-go/testing/mockprovider"
	"github.com/sensu/sensu-go/types"
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
	provider.On("Name").Return("none")
	res, err := http.Get(server.URL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestMiddlewareNoCredentials(t *testing.T) {
	provider := &mockprovider.MockProvider{}
	server := httptest.NewServer(Middleware(provider, testHandler()))
	defer server.Close()

	// No credentials passed
	provider.On("Name").Return("basic")
	res, err := http.Get(server.URL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
}

func TestMiddlewareInvalidCredentials(t *testing.T) {
	provider := &mockprovider.MockProvider{}
	server := httptest.NewServer(Middleware(provider, testHandler()))
	defer server.Close()

	// Invalid credentials
	provider.On("Name").Return("basic")
	provider.On("Authenticate").Return(&types.User{}, fmt.Errorf(""))
	client := &http.Client{}
	req, _ := http.NewRequest("GET", server.URL, nil)
	req.SetBasicAuth("foo", "P@ssw0rd!")
	res, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
}

func TestMiddlewareValidCredentials(t *testing.T) {
	provider := &mockprovider.MockProvider{}
	server := httptest.NewServer(Middleware(provider, testHandler()))
	defer server.Close()

	// Invalid credentials
	provider.On("Name").Return("basic")
	provider.On("Authenticate").Return(&types.User{}, nil)
	client := &http.Client{}
	req, _ := http.NewRequest("GET", server.URL, nil)
	req.SetBasicAuth("foo", "P@ssw0rd!")
	res, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

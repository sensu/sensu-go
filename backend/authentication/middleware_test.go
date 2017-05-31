package authentication

import (
	"fmt"
	"io/ioutil"
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

func TestMiddlewareInvalidCredentials(t *testing.T) {
	provider := &mockprovider.MockProvider{}
	server := httptest.NewServer(Middleware(provider, testHandler()))
	defer server.Close()

	// Invalid credentials
	provider.On("AuthEnabled").Return(true)
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
	server := httptest.NewServer(Middleware(provider,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// The claims should be defined in the request context
			claims := getClaimsFromContext(r)
			assert.Equal(t, "foo", claims["sub"])

			return
		}),
	))
	defer server.Close()

	// Valid credentials
	user := types.FixtureUser("foo")
	provider.On("AuthEnabled").Return(true)
	provider.On("Authenticate").Return(user, nil)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", server.URL, nil)
	req.SetBasicAuth("foo", "P@ssw0rd!")

	res, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	// The response should not contain the signed token
	body, _ := ioutil.ReadAll(res.Body)
	bodyString := string(body)
	assert.Empty(t, bodyString)
}

func TestMiddlewareLogin(t *testing.T) {
	provider := &mockprovider.MockProvider{}
	server := httptest.NewServer(Middleware(provider, testHandler()))
	defer server.Close()

	// Valid credentials
	user := types.FixtureUser("foo")
	provider.On("AuthEnabled").Return(true)
	provider.On("Authenticate").Return(user, nil)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", server.URL+"/auth", nil)
	req.SetBasicAuth("foo", "P@ssw0rd!")

	res, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	// The response should contain the signed token
	body, _ := ioutil.ReadAll(res.Body)
	bodyString := string(body)
	assert.NotEmpty(t, bodyString)
}

func TestMiddlewareJWT(t *testing.T) {
	provider := &mockprovider.MockProvider{}
	server := httptest.NewServer(Middleware(provider, testHandler()))
	defer server.Close()

	provider.On("AuthEnabled").Return(true)

	// Valid JWT
	secret = []byte("foobar")
	user := types.FixtureUser("foo")
	_, tokenString, _ := newToken(user)

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

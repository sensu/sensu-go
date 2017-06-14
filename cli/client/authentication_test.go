package client

import (
	"net/http"
	"net/http/httptest"
	"testing"

	config "github.com/sensu/sensu-go/cli/client/testing"
	"github.com/stretchr/testify/assert"

	resty "gopkg.in/resty.v0"
)

func TestCreateAccessToken(t *testing.T) {
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Method, http.MethodGet)
		assert.NotEmpty(t, r.Header["Authorization"])

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token": "foo", "expires_at": 123456789, "refresh_token": "bar"}`))
	}
	server := httptest.NewServer(http.HandlerFunc(testHandler))
	defer server.Close()

	mockConfig := &config.MockConfig{}
	restyInst := resty.New()
	client := &RestClient{resty: restyInst, config: mockConfig}

	mockConfig.On("GetString", "api-url").Return("")
	mockConfig.On("GetString", "access-token").Return("")

	token, err := client.CreateAccessToken(server.URL, "foo", "bar")
	assert.NoError(t, err)
	assert.NotNil(t, token)
}

func TestCreateAccessTokenForbidden(t *testing.T) {
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Request unauthorized", http.StatusUnauthorized)
		return
	}
	server := httptest.NewServer(http.HandlerFunc(testHandler))
	defer server.Close()

	mockConfig := &config.MockConfig{}
	restyInst := resty.New()
	client := &RestClient{resty: restyInst, config: mockConfig}

	mockConfig.On("GetString", "api-url").Return("")
	mockConfig.On("GetString", "access-token").Return("")

	_, err := client.CreateAccessToken(server.URL, "foo", "bar")
	assert.Error(t, err)
}

func TestRefreshAccessToken(t *testing.T) {
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Method, http.MethodPost)
		assert.NotEmpty(t, r.Header["Authorization"])

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token": "foo", "expires_at": 123456789, "refresh_token": "bar"}`))
	}
	server := httptest.NewServer(http.HandlerFunc(testHandler))
	defer server.Close()

	mockConfig := &config.MockConfig{}
	restyInst := resty.New()
	client := &RestClient{resty: restyInst, config: mockConfig}

	mockConfig.On("GetString", "api-url").Return(server.URL)
	mockConfig.On("GetString", "access-token").Return("foo")

	token, err := client.RefreshAccessToken("bar")
	assert.NoError(t, err)
	assert.NotNil(t, token)
}

func TestRefreshAccessTokenForbidden(t *testing.T) {
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Request unauthorized", http.StatusUnauthorized)
		return
	}
	server := httptest.NewServer(http.HandlerFunc(testHandler))
	defer server.Close()

	mockConfig := &config.MockConfig{}
	restyInst := resty.New()
	client := &RestClient{resty: restyInst, config: mockConfig}

	mockConfig.On("GetString", "api-url").Return(server.URL)
	mockConfig.On("GetString", "access-token").Return("foo")

	_, err := client.RefreshAccessToken("bar")
	assert.Error(t, err)
}

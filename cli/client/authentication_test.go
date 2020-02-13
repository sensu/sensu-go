package client

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/cli/client/config"
	"github.com/stretchr/testify/assert"
)

func TestCreateAccessToken(t *testing.T) {
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Method, http.MethodGet)
		assert.NotEmpty(t, r.Header["Authorization"])

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token": "foo", "expires_at": 123456789, "refresh_token": "bar"}`))
	}
	server := httptest.NewServer(http.HandlerFunc(testHandler))
	defer server.Close()

	mockConfig := &config.MockConfig{}
	restyInst := resty.New()
	client := &RestClient{resty: restyInst, config: mockConfig}

	mockConfig.On("APIUrl").Return("")
	mockConfig.On("Tokens").Return(&corev2.Tokens{})

	token, err := client.CreateAccessToken(server.URL, "foo", "bar")
	assert.NoError(t, err)
	assert.NotNil(t, token)
}

func TestCreateAccessTokenForbidden(t *testing.T) {
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Request unauthorized", http.StatusUnauthorized)
	}
	server := httptest.NewServer(http.HandlerFunc(testHandler))
	defer server.Close()

	mockConfig := &config.MockConfig{}
	restyInst := resty.New()
	client := &RestClient{resty: restyInst, config: mockConfig}

	mockConfig.On("APIUrl").Return("")
	mockConfig.On("Tokens").Return(&corev2.Tokens{})

	_, err := client.CreateAccessToken(server.URL, "foo", "bar")
	assert.Error(t, err)
}

func TestRefreshAccessToken(t *testing.T) {
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Method, http.MethodPost)
		assert.NotEmpty(t, r.Header["Authorization"])

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token": "foo", "expires_at": 123456789, "refresh_token": "bar"}`))
	}
	server := httptest.NewServer(http.HandlerFunc(testHandler))
	defer server.Close()

	mockConfig := &config.MockConfig{}
	restyInst := resty.New()
	client := &RestClient{resty: restyInst, config: mockConfig}

	mockConfig.On("APIUrl").Return(server.URL)
	mockConfig.On("Tokens").Return(&corev2.Tokens{Access: "foo"})

	token, err := client.RefreshAccessToken(corev2.FixtureTokens("foo", "bar"))
	assert.NoError(t, err)
	assert.NotNil(t, token)
}

func TestRefreshAccessTokenForbidden(t *testing.T) {
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Request unauthorized", http.StatusUnauthorized)
	}
	server := httptest.NewServer(http.HandlerFunc(testHandler))
	defer server.Close()

	mockConfig := &config.MockConfig{}
	restyInst := resty.New()
	client := &RestClient{resty: restyInst, config: mockConfig}

	mockConfig.On("APIUrl").Return(server.URL)
	mockConfig.On("Tokens").Return(&corev2.Tokens{Access: "foo"})

	_, err := client.RefreshAccessToken(corev2.FixtureTokens("foo", "bar"))
	assert.Error(t, err)
}

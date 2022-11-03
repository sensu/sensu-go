package client

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/cli/client/config"
	"github.com/stretchr/testify/assert"
)

func TestCreateAPIKey(t *testing.T) {
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Method, http.MethodPost)
		assert.NotEmpty(t, r.Header["Authorization"])

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`83abef1e-e7d7-4beb-91fc-79ad90084d5b`))
	}
	server := httptest.NewServer(http.HandlerFunc(testHandler))
	defer server.Close()

	mockConfig := &config.MockConfig{}
	restyInst := resty.New()
	client := &RestClient{resty: restyInst, config: mockConfig}

	mockConfig.On("APIUrl").Return(server.URL)
	mockConfig.On("Tokens").Return(&corev2.Tokens{Access: "foo"})
	mockConfig.On("APIKey").Return("")

	apikey, err := client.CreateAPIKey("user1")
	assert.NoError(t, err)
	assert.NotNil(t, apikey)
}

func TestFetchAPIKey(t *testing.T) {
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Method, http.MethodGet)
		assert.NotEmpty(t, r.Header["Authorization"])
		_, _ = w.Write([]byte(`{
"metadata": {
    "name": "83abef1e-e7d7-4beb-91fc-79ad90084d5b",
    "created_by": "user1"
  },
  "username": "user1",
  "created_at": 1570640363
}`))
	}
	server := httptest.NewServer(http.HandlerFunc(testHandler))
	defer server.Close()

	mockConfig := &config.MockConfig{}
	restyInst := resty.New()
	client := &RestClient{resty: restyInst, config: mockConfig}

	mockConfig.On("APIUrl").Return(server.URL)
	mockConfig.On("Tokens").Return(&corev2.Tokens{Access: "foo"})
	mockConfig.On("APIKey").Return("")

	apikey, err := client.FetchAPIKey("83abef1e-e7d7-4beb-91fc-79ad90084d5b")
	assert.NoError(t, err)
	assert.Equal(t, apikey.Username, "user1")
}

func TestListAPIKeys(t *testing.T) {
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Method, http.MethodGet)
		assert.NotEmpty(t, r.Header["Authorization"])
		_, _ = w.Write([]byte(`[{
"metadata": {
    "name": "83abef1e-e7d7-4beb-91fc-79ad90084d5b",
    "created_by": "admin"
  },
  "username": "user1",
  "created_at": 1570640363
},{
"metadata": {
    "name": "83abef1e-e7d7-4beb-91fc-79ad90084d5p",
    "created_by": "admin"
  },
  "username": "user2",
  "created_at": 1570640370
}]`))
	}
	server := httptest.NewServer(http.HandlerFunc(testHandler))
	defer server.Close()

	mockConfig := &config.MockConfig{}
	restyInst := resty.New()
	client := &RestClient{resty: restyInst, config: mockConfig}

	mockConfig.On("APIUrl").Return(server.URL)
	mockConfig.On("Tokens").Return(&corev2.Tokens{Access: "foo"})
	mockConfig.On("APIKey").Return("")

	var header http.Header
	var opts ListOptions
	apikeys, err := client.ListAPIKeys(&opts, &header)
	assert.NoError(t, err)
	assert.Equal(t, apikeys[0].Username, "user1")
	assert.Equal(t, apikeys[1].Username, "user2")
}

func TestUpdateAPIKey(t *testing.T) {
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Method, http.MethodPatch)
		assert.NotEmpty(t, r.Header["Authorization"])

		w.Header().Set("Content-Type", "application/merge-patch+json")
		_, _ = w.Write([]byte(`{"username": "devteam"}`))
	}
	server := httptest.NewServer(http.HandlerFunc(testHandler))
	defer server.Close()

	mockConfig := &config.MockConfig{}
	restyInst := resty.New()
	client := &RestClient{resty: restyInst, config: mockConfig}

	mockConfig.On("APIUrl").Return(server.URL)
	mockConfig.On("Tokens").Return(&corev2.Tokens{Access: "foo"})
	mockConfig.On("APIKey").Return("")

	err := client.UpdateAPIKey("83abef1e-e7d7-4beb-91fc-79ad90084d5b", "devteam")
	assert.NoError(t, err)
}

func TestDeleteAPIKey(t *testing.T) {
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Method, http.MethodDelete)
		assert.NotEmpty(t, r.Header["Authorization"])
	}
	server := httptest.NewServer(http.HandlerFunc(testHandler))
	defer server.Close()

	mockConfig := &config.MockConfig{}
	restyInst := resty.New()
	client := &RestClient{resty: restyInst, config: mockConfig}

	mockConfig.On("APIUrl").Return(server.URL)
	mockConfig.On("Tokens").Return(&corev2.Tokens{Access: "foo"})
	mockConfig.On("APIKey").Return("")

	err := client.DeleteAPIKey("83abef1e-e7d7-4beb-91fc-79ad90084d5b")
	assert.NoError(t, err)
}

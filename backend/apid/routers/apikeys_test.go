package routers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAPIKeysRouter(t *testing.T) {
	s := &mockstore.V2MockStore{}
	cs := new(mockstore.ConfigStore)
	s.On("GetConfigStore").Return(cs)
	router := NewAPIKeysRouter(s)
	parentRouter := mux.NewRouter().PathPrefix(corev2.URLPrefix).Subrouter()
	router.Mount(parentRouter)

	empty := &corev2.APIKey{}
	fixture := corev2.FixtureAPIKey("226f9e06-9d54-45c6-a9f6-4206bfa7ccf6", "bar")

	tests := []routerTestCase{}
	tests = append(tests, getTestCases[*corev2.APIKey](fixture)...)
	tests = append(tests, listTestCases[*corev2.APIKey](empty)...)
	tests = append(tests, deleteTestCases(fixture)...)
	for _, tt := range tests {
		run(t, tt, parentRouter, s)
	}
}

func TestPostAPIKey(t *testing.T) {
	s := &mockstore.V2MockStore{}
	cs := new(mockstore.ConfigStore)
	s.On("GetConfigStore").Return(cs)
	cs.On("CreateIfNotExists", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	cs.On("Get", mock.Anything, mock.Anything).Return(mockstore.Wrapper[*corev2.User]{Value: corev2.FixtureUser("admin")}, nil)
	router := NewAPIKeysRouter(s)
	parentRouter := mux.NewRouter()
	router.Mount(parentRouter)
	server := httptest.NewServer(parentRouter)
	defer server.Close()

	// Prepare the HTTP request
	fixture := corev2.FixtureAPIKey("226f9e06-9d54-45c6-a9f6-4206bfa7ccf6", "admin")
	payload, err := json.Marshal(fixture)
	assert.NoError(t, err)
	client := new(http.Client)
	req, err := http.NewRequest(http.MethodPost, server.URL+"/apikeys", bytes.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}

	// Perform the HTTP request
	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	assert.Equal(t, http.StatusCreated, res.StatusCode)
}

func TestPostAPIKeyInvalidUser(t *testing.T) {
	s := &mockstore.V2MockStore{}
	cs := new(mockstore.ConfigStore)
	s.On("GetConfigStore").Return(cs)
	var user corev2.User
	user.Username = "admin"
	cs.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	userReq := storev2.NewResourceRequestFromResource(&user)
	cs.On("Get", mock.Anything, userReq).Return(nil, &store.ErrNotFound{})
	router := NewAPIKeysRouter(s)
	parentRouter := mux.NewRouter()
	router.Mount(parentRouter)
	server := httptest.NewServer(parentRouter)
	defer server.Close()

	// Prepare the HTTP request
	fixture := corev2.FixtureAPIKey("226f9e06-9d54-45c6-a9f6-4206bfa7ccf6", "admin")
	payload, err := json.Marshal(fixture)
	assert.NoError(t, err)
	client := new(http.Client)
	req, err := http.NewRequest(http.MethodPost, server.URL+"/apikeys", bytes.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}

	// Perform the HTTP request
	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

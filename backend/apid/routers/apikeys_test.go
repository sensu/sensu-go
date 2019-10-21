package routers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAPIKeysRouter(t *testing.T) {
	s := &mockstore.MockStore{}
	s.On("GetUser", mock.Anything, mock.Anything).Return(corev2.FixtureUser("admin"), nil)
	router := NewAPIKeysRouter(s)
	parentRouter := mux.NewRouter().PathPrefix(corev2.URLPrefix).Subrouter()
	router.Mount(parentRouter)

	empty := &corev2.APIKey{}
	fixture := corev2.FixtureAPIKey("226f9e06-9d54-45c6-a9f6-4206bfa7ccf6", "bar")

	tests := []routerTestCase{}
	tests = append(tests, getTestCases(fixture)...)
	tests = append(tests, listTestCases(empty)...)
	tests = append(tests, deleteTestCases(fixture)...)
	for _, tt := range tests {
		run(t, tt, parentRouter, s)
	}
}

func TestPostAPIKey(t *testing.T) {
	s := &mockstore.MockStore{}
	s.On("CreateResource", mock.Anything, mock.Anything).Return(nil, nil)
	s.On("GetUser", mock.Anything, mock.Anything).Return(corev2.FixtureUser("admin"), nil)
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
	s := &mockstore.MockStore{}
	var user *corev2.User
	s.On("CreateResource", mock.Anything, mock.Anything).Return(nil, nil)
	s.On("GetUser", mock.Anything, mock.Anything).Return(user, nil)
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

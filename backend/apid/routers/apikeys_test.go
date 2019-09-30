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
	router := NewAPIKeysRouter(s)
	parentRouter := mux.NewRouter().PathPrefix(corev2.URLPrefix).Subrouter()
	router.Mount(parentRouter)

	empty := &corev2.APIKey{}
	fixture := corev2.FixtureAPIKey("foo", "bar")

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
	router := NewAPIKeysRouter(s)
	parentRouter := mux.NewRouter()
	router.Mount(parentRouter)
	server := httptest.NewServer(parentRouter)
	defer server.Close()

	// Prepare the HTTP request
	fixture := corev2.FixtureAPIKey("foo", "bar")
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

func TestPatchAPIKey(t *testing.T) {
	s := &mockstore.MockStore{}
	s.On("CreateOrUpdateResource", mock.Anything, mock.Anything).Return(nil, nil)
	s.On("GetResource", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
	router := NewAPIKeysRouter(s)
	parentRouter := mux.NewRouter()
	router.Mount(parentRouter)
	server := httptest.NewServer(parentRouter)
	defer server.Close()

	// Prepare the HTTP request
	fixture := corev2.FixtureAPIKey("foo", "bar")
	payload, err := json.Marshal(fixture)
	assert.NoError(t, err)
	client := new(http.Client)
	req, err := http.NewRequest(http.MethodPatch, server.URL+"/apikeys/foo", bytes.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}

	// Perform the HTTP request
	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
}

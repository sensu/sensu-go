package middlewares

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func testHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Success")
	})
}

func TestEnvironment(t *testing.T) {
	store := &mockstore.MockStore{}
	store.On(
		"GetEnvironment",
		mock.Anything,
		"foo",
		"bar",
	).Return(&types.Environment{}, nil)

	mware := Environment{Store: store}
	server := httptest.NewServer(mware.Then(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Make sure that the organization is within the request context
			org := r.Context().Value(types.OrganizationKey)
			assert.NotNil(t, org)
			orgString, _ := org.(string)
			assert.Equal(t, "foo", orgString)

			env := r.Context().Value(types.EnvironmentKey)
			assert.NotNil(t, env)
			envString, _ := env.(string)
			assert.Equal(t, "bar", envString)
		}),
	))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL, nil)
	// Add a query parameter for the organization
	query := req.URL.Query()
	query.Add("env", "bar")
	query.Add("org", "foo")
	req.URL.RawQuery = query.Encode()

	res, err := http.DefaultClient.Do(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestEnvironmentNoParameters(t *testing.T) {
	store := &mockstore.MockStore{}
	store.On(
		"GetEnvironment",
		mock.Anything,
		"default",
		"default",
	).Return(&types.Environment{}, nil)

	mware := Environment{Store: store}
	server := httptest.NewServer(mware.Then(testHandler()))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL, nil)
	res, err := http.DefaultClient.Do(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestValidateWildcard(t *testing.T) {
	store := &mockstore.MockStore{}
	store.On(
		"GetEnvironment",
		mock.Anything,
		"default",
		"default",
	).Return(&types.Environment{}, nil)

	mware := Environment{Store: store}
	server := httptest.NewServer(mware.Then(testHandler()))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL, nil)
	query := req.URL.Query()
	query.Add("env", "*")
	query.Add("org", "*")
	req.URL.RawQuery = query.Encode()

	res, err := http.DefaultClient.Do(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

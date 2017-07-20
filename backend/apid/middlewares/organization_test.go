package middlewares

import (
	"errors"
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
		return
	})
}

func TestValidateOrganization(t *testing.T) {
	store := &mockstore.MockStore{}
	store.On("GetOrganizationByName", mock.Anything, "foo").Return(&types.Organization{}, nil)

	mware := Organization{Store: store}
	server := httptest.NewServer(mware.Register(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Make sure that the organization is within the request context
			org := r.Context().Value(types.OrganizationKey)
			assert.NotNil(t, org)
			orgString, _ := org.(string)
			assert.Equal(t, "foo", orgString)

			return
		}),
	))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL, nil)
	// Add a query parameter for the organization
	query := req.URL.Query()
	query.Add("org", "foo")
	req.URL.RawQuery = query.Encode()

	res, err := http.DefaultClient.Do(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestValidateNoOrganization(t *testing.T) {
	store := &mockstore.MockStore{}

	mware := Organization{Store: store}
	server := httptest.NewServer(mware.Register(testHandler()))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL, nil)
	res, err := http.DefaultClient.Do(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestValidateOrganizationError(t *testing.T) {
	store := &mockstore.MockStore{}
	store.On(
		"GetOrganizationByName",
		mock.Anything,
		"foo",
	).Return(&types.Organization{}, errors.New("error"))

	mware := Organization{Store: store}
	server := httptest.NewServer(mware.Register(testHandler()))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL, nil)
	// Add a query parameter for the organization
	query := req.URL.Query()
	query.Add("org", "foo")
	req.URL.RawQuery = query.Encode()

	res, err := http.DefaultClient.Do(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

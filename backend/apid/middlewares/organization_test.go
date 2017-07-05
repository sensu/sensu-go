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
)

func testHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Success")
		return
	})
}

func TestValidateOrganization(t *testing.T) {
	store := &mockstore.MockStore{}
	store.On("GetOrganizationByName", "foo").Return(&types.Organization{}, nil)

	server := httptest.NewServer(Organization(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Make sure that the organization is within the request context
			org := r.Context().Value(types.OrganizationKey)
			assert.NotNil(t, org)
			orgString, _ := org.(string)
			assert.Equal(t, "foo", orgString)

			return
		}),
		store))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL, nil)
	// Add a query parameter for the organization
	query := req.URL.Query()
	query.Add("org", "foo")
	req.URL.RawQuery = query.Encode()

	client := &http.Client{}
	res, err := client.Do(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

}

func TestValidateNoOrganization(t *testing.T) {
	store := &mockstore.MockStore{}

	server := httptest.NewServer(Organization(testHandler(), store))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL, nil)
	client := &http.Client{}
	res, err := client.Do(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestValidateOrganizationError(t *testing.T) {
	store := &mockstore.MockStore{}
	store.On("GetOrganizationByName", "foo").Return(&types.Organization{}, errors.New("error"))

	server := httptest.NewServer(Organization(testHandler(), store))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL, nil)
	// Add a query parameter for the organization
	query := req.URL.Query()
	query.Add("org", "foo")
	req.URL.RawQuery = query.Encode()

	client := &http.Client{}
	res, err := client.Do(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

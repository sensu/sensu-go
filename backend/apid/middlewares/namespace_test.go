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
		_, _ = fmt.Fprint(w, "Success")
	})
}

func TestNamespace(t *testing.T) {
	store := &mockstore.MockStore{}
	store.On(
		"GetNamespace",
		mock.Anything,
		"acme",
	).Return(&types.Namespace{}, nil)

	mware := Namespace{Store: store}
	server := httptest.NewServer(mware.Then(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			namespace := r.Context().Value(types.NamespaceKey)
			assert.NotNil(t, namespace)
			namespaceString, _ := namespace.(string)
			assert.Equal(t, "acme", namespaceString)
		}),
	))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL, nil)
	// Add a query parameter for the namespace
	query := req.URL.Query()
	query.Add("namespace", "acme")
	req.URL.RawQuery = query.Encode()

	res, err := http.DefaultClient.Do(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestNamespaceNoParameters(t *testing.T) {
	store := &mockstore.MockStore{}
	store.On(
		"GetNamespace",
		mock.Anything,
		"default",
	).Return(&types.Namespace{}, nil)

	mware := Namespace{Store: store}
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
		"GetNamespace",
		mock.Anything,
		"default",
	).Return(&types.Namespace{}, nil)

	mware := Namespace{Store: store}
	server := httptest.NewServer(mware.Then(testHandler()))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL, nil)
	query := req.URL.Query()
	query.Add("namespace", types.NamespaceTypeAll)
	req.URL.RawQuery = query.Encode()

	res, err := http.DefaultClient.Do(req)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

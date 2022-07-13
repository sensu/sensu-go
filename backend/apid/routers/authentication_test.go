package routers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/authentication"
	"github.com/sensu/sensu-go/backend/authentication/providers/basic"
	realStore "github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLoginNoCredentials(t *testing.T) {
	store := &mockstore.MockStore{}
	a := authenticationRouter(store)

	req, _ := http.NewRequest(http.MethodGet, "/auth", nil)

	res := processRequest(a, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestLoginInvalidCredentials(t *testing.T) {
	store := &mockstore.MockStore{}
	a := authenticationRouter(store)

	user := types.FixtureUser("foo")
	store.
		On("AuthenticateUser", mock.Anything, "foo", "P@ssw0rd!").
		Return(user, fmt.Errorf("error"))

	req, _ := http.NewRequest(http.MethodGet, "/auth", nil)
	req.SetBasicAuth("foo", "P@ssw0rd!")

	res := processRequest(a, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestLoginSuccessful(t *testing.T) {
	store := &mockstore.MockStore{}
	a := authenticationRouter(store)

	user := types.FixtureUser("foo")
	store.
		On("AuthenticateUser", mock.Anything, "foo", "P@ssw0rd!").
		Return(user, nil)

	req, _ := http.NewRequest(http.MethodGet, "/auth", nil)
	req.SetBasicAuth("foo", "P@ssw0rd!")

	res := processRequest(a, req)
	assert.Equal(t, http.StatusOK, res.Code)

	// We should have the access token
	body := res.Body.Bytes()
	response := &types.Tokens{}
	err := json.Unmarshal(body, &response)

	assert.NoError(t, err)
	assert.NotEmpty(t, response.Access)
	assert.NotZero(t, response.ExpiresAt)
	assert.NotEmpty(t, response.Refresh)
}

func TestTestNoCredentials(t *testing.T) {
	store := &mockstore.MockStore{}
	a := authenticationRouter(store)

	req, _ := http.NewRequest(http.MethodGet, "/auth/test", nil)

	res := processRequest(a, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestTestInvalidCredentials(t *testing.T) {
	store := &mockstore.MockStore{}
	a := authenticationRouter(store)

	user := types.FixtureUser("foo")
	store.
		On("AuthenticateUser", mock.Anything, "foo", "P@ssw0rd!").
		Return(user, fmt.Errorf("error"))

	req, _ := http.NewRequest(http.MethodGet, "/auth/test", nil)
	req.SetBasicAuth("foo", "P@ssw0rd!")

	res := processRequest(a, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestTestSuccessful(t *testing.T) {
	store := &mockstore.MockStore{}
	a := authenticationRouter(store)

	user := types.FixtureUser("foo")
	store.
		On("AuthenticateUser", mock.Anything, "foo", "P@ssw0rd!").
		Return(user, nil)

	req, _ := http.NewRequest(http.MethodGet, "/auth/test", nil)
	req.SetBasicAuth("foo", "P@ssw0rd!")

	res := processRequest(a, req)
	assert.Equal(t, http.StatusOK, res.Code)
}

func authenticationRouter(store realStore.Store) *AuthenticationRouter {
	authenticator := &authentication.Authenticator{}
	provider := &basic.Provider{Store: store, ObjectMeta: corev2.ObjectMeta{Name: basic.Type}}
	authenticator.AddProvider(provider)
	return &AuthenticationRouter{authenticator: authenticator}
}

func processRequest(router Router, req *http.Request) *httptest.ResponseRecorder {
	parent := mux.NewRouter()
	router.Mount(parent)

	res := httptest.NewRecorder()
	parent.ServeHTTP(res, req)
	return res
}

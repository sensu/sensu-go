package routers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockAuthenticator struct {
	mock.Mock
}

func (m *mockAuthenticator) CreateAccessToken(ctx context.Context, username, password string) (*corev2.Tokens, error) {
	args := m.Called(ctx, username, password)
	return args.Get(0).(*corev2.Tokens), args.Error(1)
}

func (m *mockAuthenticator) TestCreds(ctx context.Context, username, password string) error {
	return m.Called(ctx, username, password).Error(0)
}

func (m *mockAuthenticator) Logout(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

func (m *mockAuthenticator) RefreshAccessToken(ctx context.Context) (*corev2.Tokens, error) {
	args := m.Called(ctx)
	return args.Get(0).(*corev2.Tokens), args.Error(1)
}

func TestLoginNoCredentials(t *testing.T) {
	auth := new(mockAuthenticator)
	router := NewAuthenticationRouter(auth)

	req, _ := http.NewRequest(http.MethodGet, "/auth", nil)

	res := processRequest(router, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestLoginInvalidCredentials(t *testing.T) {
	auth := new(mockAuthenticator)
	auth.On("CreateAccessToken", mock.Anything, mock.Anything, mock.Anything).Return((*corev2.Tokens)(nil), corev2.ErrUnauthorized)
	router := NewAuthenticationRouter(auth)

	req, _ := http.NewRequest(http.MethodGet, "/auth", nil)
	req.SetBasicAuth("foo", "P@ssw0rd!")

	res := processRequest(router, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestLoginSuccessful(t *testing.T) {
	auth := new(mockAuthenticator)
	tokens := &corev2.Tokens{
		Access:    "abcd",
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
		Refresh:   "abcd",
	}
	auth.On("CreateAccessToken", mock.Anything, "foo", "P@ssw0rd!").Return(tokens, nil)
	router := NewAuthenticationRouter(auth)

	req, _ := http.NewRequest(http.MethodGet, "/auth", nil)
	req.SetBasicAuth("foo", "P@ssw0rd!")

	res := processRequest(router, req)
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
	auth := new(mockAuthenticator)
	auth.On("TestCreds", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("not authenticated"))
	router := NewAuthenticationRouter(auth)

	req, _ := http.NewRequest(http.MethodGet, "/auth/test", nil)

	res := processRequest(router, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestTestInvalidCredentials(t *testing.T) {
	auth := new(mockAuthenticator)
	auth.On("TestCreds", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("not authenticated"))
	router := NewAuthenticationRouter(auth)

	req, _ := http.NewRequest(http.MethodGet, "/auth/test", nil)
	req.SetBasicAuth("foo", "P@ssw0rd!")

	res := processRequest(router, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestTestSuccessful(t *testing.T) {
	auth := new(mockAuthenticator)
	auth.On("TestCreds", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	router := NewAuthenticationRouter(auth)

	req, _ := http.NewRequest(http.MethodGet, "/auth/test", nil)
	req.SetBasicAuth("foo", "P@ssw0rd!")

	res := processRequest(router, req)
	assert.Equal(t, http.StatusOK, res.Code)
}

func processRequest(router Router, req *http.Request) *httptest.ResponseRecorder {
	parent := mux.NewRouter()
	router.Mount(parent)

	res := httptest.NewRecorder()
	parent.ServeHTTP(res, req)
	return res
}

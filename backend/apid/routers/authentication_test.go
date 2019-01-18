package routers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/middlewares"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/authentication/providers"
	"github.com/sensu/sensu-go/backend/authentication/providers/basic"
	"github.com/sensu/sensu-go/backend/store"
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
	store.On("AllowTokens", mock.AnythingOfType("[]*jwt.Token")).Return(nil)
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
	a := &AuthenticationRouter{store}

	req, _ := http.NewRequest(http.MethodGet, "/auth/test", nil)

	res := processRequest(a, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestTestInvalidCredentials(t *testing.T) {
	store := &mockstore.MockStore{}
	a := &AuthenticationRouter{store}

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
	a := &AuthenticationRouter{store}

	user := types.FixtureUser("foo")
	store.
		On("AuthenticateUser", mock.Anything, "foo", "P@ssw0rd!").
		Return(user, nil)

	req, _ := http.NewRequest(http.MethodGet, "/auth/test", nil)
	req.SetBasicAuth("foo", "P@ssw0rd!")

	res := processRequest(a, req)
	assert.Equal(t, http.StatusOK, res.Code)
}

func TestLogoutNotWhitelisted(t *testing.T) {
	store := &mockstore.MockStore{}
	a := authenticationRouter(store)

	// Mock calls to the store
	store.On("RevokeTokens", mock.AnythingOfType("[]*v2.Claims")).Return(fmt.Errorf("error"))

	claims := v2.FixtureClaims("foo", nil)
	_, tokenString, _ := jwt.AccessToken(claims)
	refreshClaims := &v2.Claims{StandardClaims: v2.StandardClaims(claims.Subject)}
	_, refreshTokenString, _ := jwt.RefreshToken(refreshClaims)
	body := &types.Tokens{Refresh: refreshTokenString}
	payload, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/auth/logout", bytes.NewBuffer(payload))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokenString))
	res := processRequestWithRefreshToken(a, req)

	assert.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestLogoutSuccess(t *testing.T) {
	store := &mockstore.MockStore{}
	a := authenticationRouter(store)

	// Mock calls to the store
	store.On("RevokeTokens", mock.AnythingOfType("[]*v2.Claims")).Return(nil)

	claims := v2.FixtureClaims("foo", nil)
	_, tokenString, _ := jwt.AccessToken(claims)
	refreshClaims := &v2.Claims{StandardClaims: v2.StandardClaims(claims.Subject)}
	_, refreshTokenString, _ := jwt.RefreshToken(refreshClaims)
	body := &types.Tokens{Refresh: refreshTokenString}
	payload, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/auth/logout", bytes.NewBuffer(payload))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokenString))
	res := processRequestWithRefreshToken(a, req)

	assert.Equal(t, http.StatusOK, res.Code)
}

func TestTokenRefreshTokenNotWhitelisted(t *testing.T) {
	store := &mockstore.MockStore{}
	a := authenticationRouter(store)
	user := &types.User{Username: "foo"}

	// Mock calls to the store
	store.On(
		"GetToken",
		mock.AnythingOfType("string"),
		mock.AnythingOfType("string"),
	).Return(&types.Claims{}, fmt.Errorf("error"))
	store.On("GetUser",
		mock.AnythingOfType("*context.valueCtx"),
		mock.AnythingOfType("string"),
	).Return(user, nil)

	claims := v2.FixtureClaims("foo", nil)
	_, tokenString, _ := jwt.AccessToken(claims)
	refreshClaims := &v2.Claims{StandardClaims: v2.StandardClaims(claims.Subject)}
	_, refreshTokenString, _ := jwt.RefreshToken(refreshClaims)
	body := &types.Tokens{Refresh: refreshTokenString}
	payload, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/auth/token", bytes.NewBuffer(payload))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokenString))
	res := processRequestWithRefreshToken(a, req)

	assert.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestTokenCannotWhitelistAccessToken(t *testing.T) {
	store := &mockstore.MockStore{}
	a := authenticationRouter(store)
	user := &types.User{Username: "foo"}

	// Mock calls to the store
	store.On("AllowTokens", mock.AnythingOfType("[]*jwt.Token")).Return(fmt.Errorf("error"))
	store.On("RevokeTokens", mock.AnythingOfType("[]*v2.Claims")).Return(nil)
	store.On("GetToken",
		mock.AnythingOfType("string"), mock.AnythingOfType("string"),
	).Return(&types.Claims{}, nil)
	store.On("GetUser",
		mock.AnythingOfType("*context.valueCtx"), mock.AnythingOfType("string"),
	).Return(user, nil)

	claims := v2.FixtureClaims("foo", nil)
	_, tokenString, _ := jwt.AccessToken(claims)
	refreshClaims := &v2.Claims{StandardClaims: v2.StandardClaims(claims.Subject)}
	_, refreshTokenString, _ := jwt.RefreshToken(refreshClaims)
	body := &types.Tokens{Refresh: refreshTokenString}
	payload, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/auth/token", bytes.NewBuffer(payload))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokenString))
	res := processRequestWithRefreshToken(a, req)

	assert.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestTokenSuccess(t *testing.T) {
	store := &mockstore.MockStore{}
	a := authenticationRouter(store)
	user := &types.User{Username: "foo"}

	// Mock calls to the store
	store.On("AllowTokens", mock.AnythingOfType("[]*jwt.Token")).Return(nil)
	store.On("RevokeTokens", mock.AnythingOfType("[]*v2.Claims")).Return(nil)
	store.On("GetToken",
		mock.AnythingOfType("string"), mock.AnythingOfType("string"),
	).Return(&types.Claims{}, nil)
	store.On("GetUser",
		mock.AnythingOfType("*context.valueCtx"), mock.AnythingOfType("string"),
	).Return(user, nil)

	claims := v2.FixtureClaims("foo", nil)
	_, tokenString, _ := jwt.AccessToken(claims)
	refreshClaims := &v2.Claims{StandardClaims: v2.StandardClaims(claims.Subject)}
	_, refreshTokenString, _ := jwt.RefreshToken(refreshClaims)
	body := &types.Tokens{Refresh: refreshTokenString}
	payload, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/auth/token", bytes.NewBuffer(payload))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokenString))
	res := processRequestWithRefreshToken(a, req)

	assert.Equal(t, http.StatusOK, res.Code)

	// We should have the access token
	resBody := res.Body.Bytes()
	response := &types.Tokens{}
	err := json.Unmarshal(resBody, &response)

	assert.NoError(t, err)
	assert.NotEmpty(t, response.Access)
	assert.NotEqual(t, tokenString, response.Access)
	assert.NotZero(t, response.ExpiresAt)
	assert.NotEmpty(t, response.Refresh)
}

func authenticationRouter(store store.Store) *AuthenticationRouter {
	authenticator := &providers.Authenticator{}
	provider := &basic.Provider{Store: store}
	authenticator.AddProvider(provider)
	return &AuthenticationRouter{store: store, authenticator: authenticator}
}

func processRequestWithRefreshToken(
	router Router,
	req *http.Request,
) *httptest.ResponseRecorder {
	parent := mux.NewRouter()
	router.Mount(parent)

	middleware := middlewares.RefreshToken{}
	routerStack := middleware.Then(parent)

	res := httptest.NewRecorder()
	routerStack.ServeHTTP(res, req)
	return res
}

func processRequest(router Router, req *http.Request) *httptest.ResponseRecorder {
	parent := mux.NewRouter()
	router.Mount(parent)

	res := httptest.NewRecorder()
	parent.ServeHTTP(res, req)
	return res
}

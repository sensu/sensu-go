package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLoginNoCredentials(t *testing.T) {
	store := &mockstore.MockStore{}
	a := &AuthenticationController{
		Store: store,
	}

	req, _ := http.NewRequest(http.MethodGet, "/auth", nil)

	res := processRequest(a, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestLoginInvalidCredentials(t *testing.T) {
	store := &mockstore.MockStore{}
	a := &AuthenticationController{
		Store: store,
	}

	user := types.FixtureUser("foo")
	store.On("AuthenticateUser", "foo", "P@ssw0rd!").Return(user, fmt.Errorf("Error"))

	req, _ := http.NewRequest(http.MethodGet, "/auth", nil)
	req.SetBasicAuth("foo", "P@ssw0rd!")

	res := processRequest(a, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestLoginSuccessful(t *testing.T) {
	store := &mockstore.MockStore{}
	a := &AuthenticationController{
		Store: store,
	}

	user := types.FixtureUser("foo")
	store.On("AuthenticateUser", "foo", "P@ssw0rd!").Return(user, nil)
	store.On("CreateToken", mock.AnythingOfType("*types.Claims")).Return(nil)

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

func TestTokenNoAccessToken(t *testing.T) {
	a := &AuthenticationController{}

	req, _ := http.NewRequest(http.MethodPost, "/auth/token", nil)
	res := processRequest(a, req)

	assert.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestTokenInvalidAccessToken(t *testing.T) {
	a := &AuthenticationController{}

	req, _ := http.NewRequest(http.MethodPost, "/auth/token", nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", "foobar"))
	res := processRequest(a, req)

	assert.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestTokenNoRefreshToken(t *testing.T) {
	a := &AuthenticationController{}
	_, tokenString, _ := jwt.AccessToken("foo")

	req, _ := http.NewRequest(http.MethodPost, "/auth/token", nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokenString))
	res := processRequest(a, req)
	assert.Equal(t, http.StatusBadRequest, res.Code)

	req, _ = http.NewRequest(http.MethodPost, "/auth/token", bytes.NewBuffer([]byte("foo")))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokenString))
	res = processRequest(a, req)
	assert.Equal(t, http.StatusBadRequest, res.Code)
}

func TestTokenInvalidRefreshToken(t *testing.T) {
	a := &AuthenticationController{}
	_, tokenString, _ := jwt.AccessToken("foo")
	refreshTokenString := "foobar"
	body := &types.Tokens{Refresh: refreshTokenString}
	payload, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/auth/token", bytes.NewBuffer(payload))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokenString))
	res := processRequest(a, req)

	assert.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestTokenWrongSub(t *testing.T) {
	a := &AuthenticationController{}
	_, tokenString, _ := jwt.AccessToken("foo")
	_, refreshTokenString, _ := jwt.RefreshToken("bar")
	body := &types.Tokens{Refresh: refreshTokenString}
	payload, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/auth/token", bytes.NewBuffer(payload))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokenString))
	res := processRequest(a, req)

	assert.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestTokenRefreshTokenNotWhitelisted(t *testing.T) {
	store := &mockstore.MockStore{}
	a := &AuthenticationController{
		Store: store,
	}

	// Mock calls to the store
	store.On("GetToken", mock.AnythingOfType("string")).Return(&types.Claims{}, fmt.Errorf("error"))

	_, tokenString, _ := jwt.AccessToken("foo")
	_, refreshTokenString, _ := jwt.RefreshToken("foo")
	body := &types.Tokens{Refresh: refreshTokenString}
	payload, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/auth/token", bytes.NewBuffer(payload))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokenString))
	res := processRequest(a, req)

	assert.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestTokenCannotWhitelistAccessToken(t *testing.T) {
	store := &mockstore.MockStore{}
	a := &AuthenticationController{
		Store: store,
	}

	// Mock calls to the store
	store.On("CreateToken", mock.AnythingOfType("*types.Claims")).Return(fmt.Errorf("error"))
	store.On("DeleteToken", mock.AnythingOfType("string")).Return(nil)
	store.On("GetToken", mock.AnythingOfType("string")).Return(&types.Claims{}, nil)

	_, tokenString, _ := jwt.AccessToken("foo")
	_, refreshTokenString, _ := jwt.RefreshToken("foo")
	body := &types.Tokens{Refresh: refreshTokenString}
	payload, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/auth/token", bytes.NewBuffer(payload))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokenString))
	res := processRequest(a, req)

	assert.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestTokenSuccess(t *testing.T) {
	store := &mockstore.MockStore{}
	a := &AuthenticationController{
		Store: store,
	}

	// Mock calls to the store
	store.On("CreateToken", mock.AnythingOfType("*types.Claims")).Return(nil)
	store.On("DeleteToken", mock.AnythingOfType("string")).Return(fmt.Errorf("error"))
	store.On("GetToken", mock.AnythingOfType("string")).Return(&types.Claims{}, nil)

	_, tokenString, _ := jwt.AccessToken("foo")
	_, refreshTokenString, _ := jwt.RefreshToken("foo")
	body := &types.Tokens{Refresh: refreshTokenString}
	payload, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/auth/token", bytes.NewBuffer(payload))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokenString))
	res := processRequest(a, req)

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

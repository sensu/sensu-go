package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/testing/mockprovider"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestLoginNoCredentials(t *testing.T) {
	provider := &mockprovider.MockProvider{}
	a := &AuthenticationController{
		Provider: provider,
	}

	provider.On("AuthEnabled").Return(true)

	req, _ := http.NewRequest(http.MethodGet, "/auth", nil)

	res := processRequest(a, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestLoginInvalidCredentials(t *testing.T) {
	provider := &mockprovider.MockProvider{}
	a := &AuthenticationController{
		Provider: provider,
	}

	user := types.FixtureUser("foo")
	provider.On("AuthEnabled").Return(true)
	provider.On("Authenticate").Return(user, fmt.Errorf("Error"))

	req, _ := http.NewRequest(http.MethodGet, "/auth", nil)
	req.SetBasicAuth("foo", "P@ssw0rd!")

	res := processRequest(a, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestLoginSuccessful(t *testing.T) {
	provider := &mockprovider.MockProvider{}
	a := &AuthenticationController{
		Provider: provider,
	}

	user := types.FixtureUser("foo")
	provider.On("AuthEnabled").Return(true)
	provider.On("Authenticate").Return(user, nil)

	req, _ := http.NewRequest(http.MethodGet, "/auth", nil)
	req.SetBasicAuth("foo", "P@ssw0rd!")

	res := processRequest(a, req)
	assert.Equal(t, http.StatusOK, res.Code)

	// We should have the access token
	body := res.Body.Bytes()
	response := &AuthenticationBody{}
	err := json.Unmarshal(body, &response)

	assert.NoError(t, err)
	assert.NotEmpty(t, response.AccessToken)
	assert.NotZero(t, response.ExpiresAt)
	assert.NotEmpty(t, response.RefreshToken)
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
}

func TestTokenInvalidRefreshToken(t *testing.T) {
	a := &AuthenticationController{}
	_, tokenString, _ := jwt.AccessToken("foo")
	refreshTokenString := "foobar"
	body := &AuthenticationBody{RefreshToken: refreshTokenString}
	payload, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/auth/token", bytes.NewBuffer(payload))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokenString))
	res := processRequest(a, req)

	assert.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestTokenWrongSub(t *testing.T) {
	a := &AuthenticationController{}
	_, tokenString, _ := jwt.AccessToken("foo")
	refreshTokenString, _ := jwt.RefreshToken("bar")
	body := &AuthenticationBody{RefreshToken: refreshTokenString}
	payload, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/auth/token", bytes.NewBuffer(payload))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokenString))
	res := processRequest(a, req)

	assert.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestTokenSuccess(t *testing.T) {
	a := &AuthenticationController{}
	_, tokenString, _ := jwt.AccessToken("foo")
	refreshTokenString, _ := jwt.RefreshToken("foo")
	body := &AuthenticationBody{RefreshToken: refreshTokenString}
	payload, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/auth/token", bytes.NewBuffer(payload))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokenString))
	res := processRequest(a, req)

	assert.Equal(t, http.StatusOK, res.Code)

	// We should have the access token
	resBody := res.Body.Bytes()
	response := &AuthenticationBody{}
	err := json.Unmarshal(resBody, &response)

	assert.NoError(t, err)
	assert.NotEmpty(t, response.AccessToken)
	assert.NotEqual(t, tokenString, response.AccessToken)
	assert.NotZero(t, response.ExpiresAt)
	assert.NotEmpty(t, response.RefreshToken)
}

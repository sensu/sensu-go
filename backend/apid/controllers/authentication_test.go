package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

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
	response := &authenticationSuccessResponse{}
	err := json.Unmarshal(body, &response)

	assert.NoError(t, err)
	assert.NotEmpty(t, response.AccessToken)
	// assert.NotZero(t, response.ExpiresAt) # Expiration not activated yet
}

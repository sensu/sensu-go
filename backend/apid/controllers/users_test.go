package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/sensu/sensu-go/testing/mockprovider"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestHttpApiUsersPut(t *testing.T) {
	provider := &mockprovider.MockProvider{}
	u := &UsersController{
		Authentication: provider,
	}

	provider.On("CreateUser").Return(nil)

	user := types.FixtureUser("foo")
	user.Password = "P@ssw0rd!"
	userBytes, _ := json.Marshal(user)

	req, _ := http.NewRequest("PUT", fmt.Sprintf("/users"), bytes.NewBuffer(userBytes))
	res := processRequest(u, req)

	assert.Equal(t, http.StatusCreated, res.Code)
}

func TestHttpApiUsersPutError(t *testing.T) {
	provider := &mockprovider.MockProvider{}
	u := &UsersController{
		Authentication: provider,
	}

	provider.On("CreateUser").Return(fmt.Errorf(""))

	user := types.FixtureUser("foo")
	user.Password = "P@ssw0rd!"
	userBytes, _ := json.Marshal(user)

	req, _ := http.NewRequest("PUT", fmt.Sprintf("/users"), bytes.NewBuffer(userBytes))
	res := processRequest(u, req)

	assert.Equal(t, http.StatusInternalServerError, res.Code)
}

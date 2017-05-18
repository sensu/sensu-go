package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

type mockProvider struct{}

func (m *mockProvider) CreateUser(u *types.User) error {
	return nil
}

func TestHttpApiUsersPut(t *testing.T) {
	mock := &mockProvider{}
	u := &UsersController{
		Authentication: mock,
	}

	user := types.FixtureUser("foo")
	user.Password = "P@ssw0rd!"
	userBytes, _ := json.Marshal(user)

	req, _ := http.NewRequest("PUT", fmt.Sprintf("/users"), bytes.NewBuffer(userBytes))
	res := processRequest(u, req)

	assert.Equal(t, http.StatusCreated, res.Code)
}

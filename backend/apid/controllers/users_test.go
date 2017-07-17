package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeleteUser(t *testing.T) {
	store := &mockstore.MockStore{}
	u := &UsersController{
		Store: store,
	}

	store.On("DeleteUserByName", "foo").Return(nil)
	req, _ := http.NewRequest(http.MethodDelete, "/rbac/users/foo", nil)
	res := processRequest(u, req)

	assert.Equal(t, http.StatusAccepted, res.Code)

	// Invalid user
	store.On("DeleteUserByName", "bar").Return(fmt.Errorf(""))
	req, _ = http.NewRequest(http.MethodDelete, "/rbac/users/bar", nil)
	res = processRequest(u, req)

	assert.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestMany(t *testing.T) {
	store := &mockstore.MockStore{}

	u := &UsersController{
		Store: store,
	}

	user1 := types.FixtureUser("foo")
	user1.Password = "P@ssw0rd!"
	user2 := types.FixtureUser("bar")

	users := []*types.User{
		user1,
		user2,
	}
	store.On("GetUsers").Return(users, nil)
	req, _ := http.NewRequest("GET", "/rbac/users", nil)
	res := processRequest(u, req)

	assert.Equal(t, http.StatusOK, res.Code)

	body := res.Body.Bytes()

	returnedUsers := []*types.User{}
	err := json.Unmarshal(body, &returnedUsers)

	assert.NoError(t, err)
	assert.EqualValues(t, users, returnedUsers)

	// The users passwords should be obfuscated
	assert.Empty(t, returnedUsers[0].Password)
}

func TestManyError(t *testing.T) {
	store := &mockstore.MockStore{}

	u := &UsersController{
		Store: store,
	}

	users := []*types.User{}
	store.On("GetUsers").Return(users, errors.New("error"))
	req, _ := http.NewRequest("GET", "/rbac/users", nil)
	res := processRequest(u, req)

	body := res.Body.Bytes()

	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.Equal(t, "error\n", string(body))
}

func TestSingle(t *testing.T) {
	store := &mockstore.MockStore{}

	u := &UsersController{
		Store: store,
	}

	var nilUser *types.User
	store.On("GetUser", "foo").Return(nilUser, nil)
	req, _ := http.NewRequest("GET", "/rbac/users/foo", nil)
	res := processRequest(u, req)

	assert.Equal(t, http.StatusNotFound, res.Code)

	user := types.FixtureUser("bar")
	user.Password = "P@ssw0rd!"
	store.On("GetUser", "bar").Return(user, nil)
	req, _ = http.NewRequest("GET", "/rbac/users/bar", nil)
	res = processRequest(u, req)

	assert.Equal(t, http.StatusOK, res.Code)

	body := res.Body.Bytes()
	result := &types.User{}
	err := json.Unmarshal(body, &result)

	assert.NoError(t, err)
	assert.Equal(t, result.Username, result.Username)

	// The user password should be obfuscated
	assert.Empty(t, result.Password)
}

func TestUpdateUser(t *testing.T) {
	store := &mockstore.MockStore{}
	u := &UsersController{
		Store: store,
	}

	storedRoles := []*types.Role{
		{Name: "default"},
	}

	user := types.FixtureUser("foo")
	user.Password = "P@ssw0rd!"
	userBytes, _ := json.Marshal(user)

	store.On("GetRoles").Return(storedRoles, nil)
	store.On("CreateUser", mock.AnythingOfType("*types.User")).Return(nil)

	req, _ := http.NewRequest("PUT", fmt.Sprintf("/rbac/users"), bytes.NewBuffer(userBytes))
	res := processRequest(u, req)

	assert.Equal(t, http.StatusCreated, res.Code)
}

func TestUpdateUserError(t *testing.T) {
	store := &mockstore.MockStore{}
	u := &UsersController{
		Store: store,
	}

	storedRoles := []*types.Role{
		{Name: "default"},
	}

	user := types.FixtureUser("foo")
	user.Password = "P@ssw0rd!"
	userBytes, _ := json.Marshal(user)

	store.On("GetRoles").Return(storedRoles, nil)
	store.On("CreateUser", mock.AnythingOfType("*types.User")).Return(fmt.Errorf(""))

	req, _ := http.NewRequest("PUT", fmt.Sprintf("/rbac/users"), bytes.NewBuffer(userBytes))
	res := processRequest(u, req)

	assert.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestValidateRoles(t *testing.T) {
	store := &mockstore.MockStore{}

	roles := []string{"roleOne", "roleTwo"}

	storedRoles := []*types.Role{
		{Name: "roleOne"},
		{Name: "roleTwo"},
	}

	store.On("GetRoles").Return(storedRoles, nil)

	assert.NoError(t, validateRoles(store, roles))
}

func TestValidateRolesError(t *testing.T) {
	store := &mockstore.MockStore{}
	roles := []string{"roleOne", "roleTwo"}

	storedRoles := []*types.Role{
		{Name: "roleOne"},
	}

	store.On("GetRoles").Return(storedRoles, nil)

	assert.Error(t, validateRoles(store, roles))
}

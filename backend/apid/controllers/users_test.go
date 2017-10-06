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
	store.On("DeleteTokensByUsername", "foo").Return(nil)

	req := newRequest(http.MethodDelete, "/rbac/users/foo", nil)
	res := processRequest(u, req)
	assert.Equal(t, http.StatusOK, res.Code)

	// Invalid user
	store.On("DeleteUserByName", "bar").Return(fmt.Errorf("error"))
	req = newRequest(http.MethodDelete, "/rbac/users/bar", nil)
	res = processRequest(u, req)
	assert.Equal(t, http.StatusInternalServerError, res.Code)

	// Unable to delete the tokens
	store.On("DeleteUserByName", "foo").Return(nil)
	store.On("DeleteTokensByUsername", "foo").Return(fmt.Errorf("error"))
	req = newRequest(http.MethodDelete, "/rbac/users/bar", nil)
	res = processRequest(u, req)
	assert.Equal(t, http.StatusInternalServerError, res.Code)

	// Unauthorized user
	req = newRequest(http.MethodDelete, "/rbac/users/bar", nil)
	req = requestWithNoAccess(req)
	res = processRequest(u, req)

	assert.Equal(t, http.StatusUnauthorized, res.Code)
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
	store.On("GetAllUsers").Return(users, nil)
	req := newRequest("GET", "/rbac/users", nil)
	res := processRequest(u, req)

	assert.Equal(t, http.StatusOK, res.Code)

	body := res.Body.Bytes()

	returnedUsers := []*types.User{}
	err := json.Unmarshal(body, &returnedUsers)

	assert.NoError(t, err)
	assert.EqualValues(t, users, returnedUsers)

	// The users passwords should be obfuscated
	assert.Empty(t, returnedUsers[0].Password)

	// Unauthorized user
	req = newRequest(http.MethodGet, "/rbac/users", nil)
	req = requestWithNoAccess(req)
	res = processRequest(u, req)
	assert.Equal(t, http.StatusOK, res.Code)

	unauthUsers := []*types.User{}
	err = json.Unmarshal(res.Body.Bytes(), &unauthUsers)
	assert.NoError(t, err)
	assert.Empty(t, unauthUsers)
}

func TestManyError(t *testing.T) {
	store := &mockstore.MockStore{}

	u := &UsersController{
		Store: store,
	}

	users := []*types.User{}
	store.On("GetAllUsers").Return(users, errors.New("error"))
	req := newRequest("GET", "/rbac/users", nil)
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
	req := newRequest("GET", "/rbac/users/foo", nil)
	res := processRequest(u, req)

	assert.Equal(t, http.StatusNotFound, res.Code)

	user := types.FixtureUser("bar")
	user.Password = "P@ssw0rd!"
	store.On("GetUser", "bar").Return(user, nil)
	req = newRequest("GET", "/rbac/users/bar", nil)
	res = processRequest(u, req)

	assert.Equal(t, http.StatusOK, res.Code)

	body := res.Body.Bytes()
	result := &types.User{}
	err := json.Unmarshal(body, &result)

	assert.NoError(t, err)
	assert.Equal(t, result.Username, result.Username)

	// The user password should be obfuscated
	assert.Empty(t, result.Password)

	// Unauthorized user
	req = newRequest(http.MethodGet, "/rbac/users/bar", nil)
	req = requestWithNoAccess(req)
	res = processRequest(u, req)

	assert.Equal(t, http.StatusUnauthorized, res.Code)
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
	store.On("GetUser", "foo").Return(user, nil)
	store.On("CreateUser", mock.AnythingOfType("*types.User")).Return(nil)

	req := newRequest("PUT", fmt.Sprintf("/rbac/users"), bytes.NewBuffer(userBytes))
	res := processRequest(u, req)

	assert.Equal(t, http.StatusCreated, res.Code)

	// Unauthorized user
	req = newRequest(http.MethodPut, "/rbac/users", bytes.NewBuffer(userBytes))
	req = requestWithNoAccess(req)
	res = processRequest(u, req)

	assert.Equal(t, http.StatusUnauthorized, res.Code)
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
	store.On("GetUser", "foo").Return(user, nil)
	store.On("CreateUser", mock.AnythingOfType("*types.User")).Return(fmt.Errorf(""))

	req := newRequest("PUT", fmt.Sprintf("/rbac/users"), bytes.NewBuffer(userBytes))
	res := processRequest(u, req)

	assert.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestUpdatePassword(t *testing.T) {
	store := &mockstore.MockStore{}
	u := &UsersController{Store: store}

	user := types.FixtureUser("foo")
	params := map[string]string{"password": "Meowmix#123"}
	paramsBytes, _ := json.Marshal(params)

	store.On("GetUser", "foo").Return(user, nil)
	store.On("UpdateUser", mock.AnythingOfType("*types.User")).Return(nil)

	req := newRequest("PUT", fmt.Sprintf("/rbac/users/foo/password"), bytes.NewBuffer(paramsBytes))
	res := processRequest(u, req)

	assert.Equal(t, http.StatusOK, res.Code)

	// Bad body
	req = newRequest(
		"PUT",
		"/rbac/users/foo/password",
		bytes.NewBuffer([]byte("ksajdf")),
	)

	res = processRequest(u, req)
	assert.Equal(t, http.StatusBadRequest, res.Code)

	// Unauthorized user
	req = newRequest(http.MethodPut, "/rbac/users/foo/password", bytes.NewBuffer(paramsBytes))
	req = requestWithNoAccess(req)
	res = processRequest(u, req)

	assert.Equal(t, http.StatusUnauthorized, res.Code)

	// Bad password
	params = map[string]string{"password": "123"}
	paramsBytes, _ = json.Marshal(params)

	req = newRequest(
		"PUT",
		"/rbac/users/foo/password",
		bytes.NewBuffer(paramsBytes),
	)
	res = processRequest(u, req)
	assert.Equal(t, http.StatusUnprocessableEntity, res.Code)

	// Bad response from store
	store.On("GetUser", "foo2").Return(user, errors.New("test"))
	req = newRequest(
		"PUT",
		"/rbac/users/foo2/password",
		bytes.NewBuffer(paramsBytes),
	)

	res = processRequest(u, req)
	assert.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestReinstateUser(t *testing.T) {
	store := &mockstore.MockStore{}
	u := &UsersController{Store: store}
	user := types.FixtureUser("foo")

	store.On("GetUser", "foo").Return(user, nil)
	store.On("UpdateUser", mock.AnythingOfType("*types.User")).Return(nil)

	req := newRequest("PUT", fmt.Sprintf("/rbac/users/foo/reinstate"), nil)
	res := processRequest(u, req)

	assert.Equal(t, http.StatusOK, res.Code)

	// Unauthorized user
	req = newRequest(http.MethodPut, "/rbac/users/foo/reinstate", nil)
	req = requestWithNoAccess(req)
	res = processRequest(u, req)

	assert.Equal(t, http.StatusUnauthorized, res.Code)

	// Bad response from store
	store.On("GetUser", "foo2").Return(user, errors.New("test"))
	req = newRequest("PUT", "/rbac/users/foo2/reinstate", nil)

	res = processRequest(u, req)
	assert.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestUserAddRole(t *testing.T) {
	store := &mockstore.MockStore{}
	u := &UsersController{Store: store}

	user := types.FixtureUser("foo")
	user.Roles = []string{}

	store.On("GetRoles").Return([]*types.Role{{Name: "admin"}}, nil)
	store.On("GetUser", "foo").Return(user, nil)
	store.On("UpdateUser", mock.AnythingOfType("*types.User")).Return(nil)

	req := newRequest("PUT", fmt.Sprintf("/rbac/users/foo/roles/admin"), nil)
	res := processRequest(u, req)

	assert.Equal(t, http.StatusOK, res.Code)

	// Unauthorized user
	req = newRequest(http.MethodPut, "/rbac/users/foo/roles/admin", nil)
	req = requestWithNoAccess(req)
	res = processRequest(u, req)

	assert.Equal(t, http.StatusUnauthorized, res.Code)

	// Invalid Role
	req = newRequest(http.MethodPut, "/rbac/users/foo/roles/asdfasdfsa", nil)
	res = processRequest(u, req)

	assert.Equal(t, http.StatusUnprocessableEntity, res.Code)

	// Bad response from store
	store.On("GetUser", "foo2").Return(user, errors.New("test"))
	req = newRequest("PUT", "/rbac/users/foo2/roles/admin", nil)

	res = processRequest(u, req)
	assert.Equal(t, http.StatusInternalServerError, res.Code)

	// Bad response from store on save
	store.On("GetUser", "foo2").Return(user, nil)
	store.On("UpdateUser", mock.AnythingOfType("*types.User")).Return(errors.New("test"))
	req = newRequest("PUT", "/rbac/users/foo2/roles/admin", nil)

	res = processRequest(u, req)
	assert.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestUserRemoveRole(t *testing.T) {
	store := &mockstore.MockStore{}
	u := &UsersController{Store: store}

	user := types.FixtureUser("foo")
	user.Roles = []string{}

	store.On("GetRoles").Return([]*types.Role{{Name: "admin"}}, nil)
	store.On("GetUser", "foo").Return(user, nil)
	store.On("UpdateUser", mock.AnythingOfType("*types.User")).Return(nil)

	req := newRequest(http.MethodDelete, fmt.Sprintf("/rbac/users/foo/roles/admin"), nil)
	res := processRequest(u, req)

	assert.Equal(t, http.StatusOK, res.Code)

	// Unauthorized user
	req = newRequest(http.MethodDelete, "/rbac/users/foo/roles/admin", nil)
	req = requestWithNoAccess(req)
	res = processRequest(u, req)

	assert.Equal(t, http.StatusUnauthorized, res.Code)

	// Bad response from store
	store.On("GetUser", "foo2").Return(user, errors.New("test"))
	req = newRequest(http.MethodDelete, "/rbac/users/foo2/roles/admin", nil)

	res = processRequest(u, req)
	assert.Equal(t, http.StatusInternalServerError, res.Code)

	// Bad response from store on save
	store.On("GetUser", "foo2").Return(user, nil)
	store.On("UpdateUser", mock.AnythingOfType("*types.User")).Return(errors.New("test"))
	req = newRequest(http.MethodDelete, "/rbac/users/foo2/roles/admin", nil)

	res = processRequest(u, req)
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

	// Single role missing
	store.On("GetRoles").Return(storedRoles, nil)
	err := validateRoles(store, roles)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "given role")

	// Multiple roles missing
	roles = append(roles, "roleThree")
	err = validateRoles(store, roles)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "given roles")
}

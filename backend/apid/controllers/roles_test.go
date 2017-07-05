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
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type RolesControllerSuite struct {
	suite.Suite

	store      *mockstore.MockStore
	controller *RolesController
}

func (suite *RolesControllerSuite) SetupTest() {
	store := &mockstore.MockStore{}

	suite.store = store
	suite.controller = &RolesController{Store: store}
}

func (suite *RolesControllerSuite) TestGetRoles() {
	roles := []*types.Role{
		types.FixtureRole("bob", "builders"),
		types.FixtureRole("fred", "builders"),
	}
	suite.store.On("GetRoles").Return(roles, nil)

	req, _ := http.NewRequest("GET", "/rbac/roles", nil)
	res := processRequest(suite.controller, req)

	suite.Equal(http.StatusOK, res.Code)

	body := res.Body.Bytes()
	receivedRecords := []*types.Role{}
	err := json.Unmarshal(body, &receivedRecords)

	suite.NoError(err)
	suite.Len(receivedRecords, 2)
	for i, role := range receivedRecords {
		suite.EqualValues(roles[i], role)
	}
}

func (suite *RolesControllerSuite) TestGetRolesWithStoreError() {
	roles := []*types.Role{}
	suite.store.On("GetRoles").Return(roles, fmt.Errorf(""))

	req, _ := http.NewRequest("GET", "/rbac/roles", nil)
	res := processRequest(suite.controller, req)

	suite.Equal(http.StatusInternalServerError, res.Code)
}

func (suite *RolesControllerSuite) TestGetRoleWithStoreError() {
	rName := "bob"
	suite.store.On("GetRoleByName", rName).Return(&types.Role{}, fmt.Errorf(""))

	req, _ := http.NewRequest("GET", "/rbac/roles/"+rName, nil)
	res := processRequest(suite.controller, req)

	suite.Equal(http.StatusInternalServerError, res.Code)
}

func (suite *RolesControllerSuite) TestGeRoleNotFound() {
	rName := "bob"
	suite.store.On("GetRoleByName", rName).Return(nil, nil)

	req, _ := http.NewRequest("GET", "/rbac/roles/"+rName, nil)
	res := processRequest(suite.controller, req)

	suite.Equal(http.StatusNotFound, res.Code)
}

func (suite *RolesControllerSuite) TestGetRole() {
	name := "bob"
	role := &types.Role{}
	suite.store.On("GetRoleByName", name).Return(role, nil)

	req, _ := http.NewRequest("GET", "/rbac/roles/"+name, nil)
	res := processRequest(suite.controller, req)

	suite.Equal(http.StatusOK, res.Code)
	suite.NotEmpty(res.Body.Bytes())
}

func (suite *RolesControllerSuite) TestCreateRoleWithError() {
	name := "bob"
	role := types.FixtureRole("name", "org")
	roleJSON, _ := json.Marshal(role)

	suite.store.On("UpdateRole", mock.AnythingOfType("*types.Role")).Return(fmt.Errorf(""))

	req, _ := http.NewRequest("PUT", "/rbac/roles/"+name, bytes.NewBuffer(roleJSON))
	res := processRequest(suite.controller, req)

	suite.Equal(http.StatusInternalServerError, res.Code)
}

func (suite *RolesControllerSuite) TestCreateRoleWithBadData() {
	name := "bob"
	roleBytes := bytes.NewBuffer([]byte("kasjdlfkajs;dlf"))
	req, _ := http.NewRequest("PUT", "/rbac/roles/"+name, roleBytes)

	res := processRequest(suite.controller, req)
	suite.Equal(http.StatusBadRequest, res.Code)
}

func (suite *RolesControllerSuite) TestCreateRoleWithInvalidRole() {
	name := "bob"
	role := types.FixtureRole("name", "org")
	role.Name = "Really;Bad--Invalid--!!!Name"

	roleJSON, _ := json.Marshal(role)
	req, _ := http.NewRequest("PUT", "/rbac/roles/"+name, bytes.NewBuffer(roleJSON))

	res := processRequest(suite.controller, req)
	suite.Equal(http.StatusBadRequest, res.Code)
}

func (suite *RolesControllerSuite) TestCreateRole() {
	name := "bob"
	role := types.FixtureRole(name, "my-org")

	roleJSON, _ := json.Marshal(role)
	req, _ := http.NewRequest("PUT", "/rbac/roles/"+name, bytes.NewBuffer(roleJSON))

	suite.store.On("UpdateRole", mock.AnythingOfType("*types.Role")).Return(nil)

	res := processRequest(suite.controller, req)
	suite.Equal(http.StatusOK, res.Code)
	suite.Empty(res.Body.Bytes())
}

func (suite *RolesControllerSuite) TestDeleteRoleWithStoreError() {
	name := "bob"
	role := &types.Role{}
	req, _ := http.NewRequest("DELETE", "/rbac/roles/"+name, nil)

	suite.store.On("GetRoleByName", name).Return(role, nil)
	suite.store.On("DeleteRoleByName", name).Return(errors.New(""))

	res := processRequest(suite.controller, req)
	suite.Equal(http.StatusInternalServerError, res.Code)
}

func (suite *RolesControllerSuite) TestDeleteRole() {
	name := "bob"
	role := &types.Role{}
	req, _ := http.NewRequest("DELETE", "/rbac/roles/"+name, nil)

	suite.store.On("GetRoleByName", name).Return(role, nil)
	suite.store.On("DeleteRoleByName", name).Return(nil)

	res := processRequest(suite.controller, req)
	suite.Equal(http.StatusOK, res.Code)
	suite.Empty(res.Body.Bytes())
}

func TestRolesController(t *testing.T) {
	suite.Run(t, new(RolesControllerSuite))
}

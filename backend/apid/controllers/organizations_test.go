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

func TestCreateOrg(t *testing.T) {
	store := &mockstore.MockStore{}
	controller := &OrganizationsController{
		Store: store,
	}

	store.On("UpdateOrganization", mock.AnythingOfType("*types.Organization")).Return(nil)

	org := types.FixtureOrganization("foo")
	orgBytes, _ := json.Marshal(org)

	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/rbac/organizations"), bytes.NewBuffer(orgBytes))
	res := processRequest(controller, req)

	assert.Equal(t, http.StatusCreated, res.Code)
}

func TestCreateOrgError(t *testing.T) {
	store := &mockstore.MockStore{}
	controller := &OrganizationsController{
		Store: store,
	}

	store.On("UpdateOrganization", mock.AnythingOfType("*types.Organization")).Return(fmt.Errorf("error"))

	org := types.FixtureOrganization("foo")
	orgBytes, _ := json.Marshal(org)

	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/rbac/organizations"), bytes.NewBuffer(orgBytes))
	res := processRequest(controller, req)

	assert.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestDeleteOrg(t *testing.T) {
	store := &mockstore.MockStore{}
	controller := &OrganizationsController{
		Store: store,
	}

	store.On("DeleteOrganizationByName", "foo").Return(nil)
	req, _ := http.NewRequest(http.MethodDelete, "/rbac/organizations/foo", nil)
	res := processRequest(controller, req)

	assert.Equal(t, http.StatusAccepted, res.Code)

	// Invalid org
	store.On("DeleteOrganizationByName", "bar").Return(fmt.Errorf(""))
	req, _ = http.NewRequest(http.MethodDelete, "/rbac/organizations/bar", nil)
	res = processRequest(controller, req)

	assert.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestManyOrg(t *testing.T) {
	store := &mockstore.MockStore{}

	controller := &OrganizationsController{
		Store: store,
	}

	org1 := types.FixtureOrganization("foo")
	org2 := types.FixtureOrganization("bar")

	orgs := []*types.Organization{
		org1,
		org2,
	}
	store.On("GetOrganizations").Return(orgs, nil)
	req, _ := http.NewRequest("GET", "/rbac/organizations", nil)
	res := processRequest(controller, req)

	assert.Equal(t, http.StatusOK, res.Code)

	body := res.Body.Bytes()

	result := []*types.Organization{}
	err := json.Unmarshal(body, &result)

	assert.NoError(t, err)
	assert.EqualValues(t, orgs, result)
}

func TestManyOrgError(t *testing.T) {
	store := &mockstore.MockStore{}

	controller := &OrganizationsController{
		Store: store,
	}

	orgs := []*types.Organization{}
	store.On("GetOrganizations").Return(orgs, errors.New("error"))
	req, _ := http.NewRequest("GET", "/rbac/organizations", nil)
	res := processRequest(controller, req)

	body := res.Body.Bytes()

	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.Equal(t, "error\n", string(body))
}

func TestSingleOrg(t *testing.T) {
	store := &mockstore.MockStore{}

	controller := &OrganizationsController{
		Store: store,
	}

	var nilOrg *types.Organization
	store.On("GetOrganizationByName", "foo").Return(nilOrg, nil)
	req, _ := http.NewRequest("GET", "/rbac/organizations/foo", nil)
	res := processRequest(controller, req)

	assert.Equal(t, http.StatusNotFound, res.Code)

	org := types.FixtureOrganization("bar")
	store.On("GetOrganizationByName", "bar").Return(org, nil)
	req, _ = http.NewRequest("GET", "/rbac/organizations/bar", nil)
	res = processRequest(controller, req)

	assert.Equal(t, http.StatusOK, res.Code)

	body := res.Body.Bytes()
	result := &types.Organization{}
	err := json.Unmarshal(body, &result)

	assert.NoError(t, err)
	assert.Equal(t, org.Name, result.Name)
}

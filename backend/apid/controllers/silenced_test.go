package controllers

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHttpApiSilencedGet(t *testing.T) {
	store := &mockstore.MockStore{}

	c := &SilencedController{
		Store: store,
	}

	silenced := []*types.Silenced{
		types.FixtureSilenced("check1"),
		types.FixtureSilenced("check2"),
	}
	store.On("GetSilencedEntries", mock.Anything).Return(silenced, nil)
	req := newRequest("GET", "/silenced", nil)
	res := processRequest(c, req)

	assert.Equal(t, http.StatusOK, res.Code)

	body := res.Body.Bytes()

	returnedSilencedEntries := []*types.Silenced{}
	err := json.Unmarshal(body, &returnedSilencedEntries)

	assert.NoError(t, err)
	assert.EqualValues(t, silenced, returnedSilencedEntries)
}

func TestHttpApiSilencedPost(t *testing.T) {
}

func TestHttpApiSilencedClear(t *testing.T) {
}

/*
func TestHttpApiChecksGetPercent(t *testing.T) {
	store := &mockstore.MockStore{}

	c := &ChecksController{
		Store: store,
	}

	checkConfig := types.FixtureCheckConfig("check1")
	checkConfig.Command = "/usr/lib/nagios/plugins/check_ping -H 192.168.250.123 -w 50,5% -c 100,10%"
	checks := []*types.CheckConfig{
		checkConfig,
	}
	store.On("GetCheckConfigs", mock.Anything).Return(checks, nil)
	req := newRequest("GET", "/checks", nil)
	res := processRequest(c, req)

	assert.Equal(t, http.StatusOK, res.Code)

	body := res.Body.Bytes()

	returnedChecks := []*types.CheckConfig{}
	err := json.Unmarshal(body, &returnedChecks)

	assert.NoError(t, err)
	assert.EqualValues(t, checks, returnedChecks)
}

func TestHttpApiChecksGetError(t *testing.T) {
	store := &mockstore.MockStore{}

	c := &ChecksController{
		Store: store,
	}

	var nilChecks []*types.CheckConfig
	store.On("GetCheckConfigs", mock.Anything).Return(nilChecks, errors.New("error"))
	req := newRequest("GET", "/checks", nil)
	res := processRequest(c, req)

	body := res.Body.Bytes()

	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.Equal(t, "error\n", string(body))
}

func TestHttpApiChecksGetUnauthorized(t *testing.T) {
	controller := ChecksController{}

	req := newRequest("GET", "/checks", nil)
	req = requestWithNoAccess(req)

	res := processRequest(&controller, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestHttpApiCheckGet(t *testing.T) {
	store := &mockstore.MockStore{}

	c := &ChecksController{
		Store: store,
	}

	var nilCheck *types.CheckConfig
	store.On("GetCheckConfigByName", mock.Anything, "somecheck").Return(nilCheck, nil)
	notFoundReq := newRequest("GET", "/checks/somecheck", nil)
	notFoundRes := processRequest(c, notFoundReq)

	assert.Equal(t, http.StatusNotFound, notFoundRes.Code)

	check1 := types.FixtureCheckConfig("check1")
	store.On("GetCheckConfigByName", mock.Anything, "check1").Return(check1, nil)
	foundReq := newRequest("GET", "/checks/check1", nil)
	foundRes := processRequest(c, foundReq)

	assert.Equal(t, http.StatusOK, foundRes.Code)

	body := foundRes.Body.Bytes()

	check := &types.CheckConfig{}
	err := json.Unmarshal(body, &check)

	assert.NoError(t, err)
	assert.NotNil(t, check.Name)
	assert.NotNil(t, check.Command)
	assert.NotEqual(t, check.Name, "")
	assert.NotEqual(t, check.Command, "")
}

func TestHttpApiChecksGetByNameUnauthorized(t *testing.T) {
	store := &mockstore.MockStore{}
	controller := ChecksController{Store: store}

	check := types.FixtureCheckConfig("nil")
	store.On("GetCheckConfigByName", mock.Anything, "nil").Return(check, nil)

	req := newRequest("GET", "/checks/nil", nil)
	req = requestWithNoAccess(req)

	res := processRequest(&controller, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestHttpApiCheckPut(t *testing.T) {
	store := &mockstore.MockStore{}

	c := &ChecksController{
		Store: store,
	}

	check := types.FixtureCheckConfig("check1")
	updatedCheckJSON, _ := json.Marshal(check)

	store.On("UpdateCheckConfig", mock.Anything, mock.AnythingOfType("*types.CheckConfig")).Return(nil).Run(func(args mock.Arguments) {
		receivedCheck := args.Get(1).(*types.CheckConfig)
		assert.NoError(t, receivedCheck.Validate())
		assert.EqualValues(t, check, receivedCheck)
	})
	putReq := newRequest("PUT", fmt.Sprintf("/checks/%s", "check1"), bytes.NewBuffer(updatedCheckJSON))
	putRes := processRequest(c, putReq)

	assert.Equal(t, http.StatusOK, putRes.Code)

	unauthReq := newRequest("PUT", "/checks/check1", bytes.NewBuffer(updatedCheckJSON))
	unauthReq = requestWithNoAccess(unauthReq)

	unauthRes := processRequest(c, unauthReq)
	assert.Equal(t, http.StatusUnauthorized, unauthRes.Code)
}

func TestHttpApiCheckPost(t *testing.T) {
	store := &mockstore.MockStore{}

	c := &ChecksController{
		Store: store,
	}

	check := types.FixtureCheckConfig("check1")
	updatedCheckJSON, _ := json.Marshal(check)

	store.On("UpdateCheckConfig", mock.Anything, mock.AnythingOfType("*types.CheckConfig")).Return(nil).Run(func(args mock.Arguments) {
		receivedCheck := args.Get(1).(*types.CheckConfig)
		assert.NoError(t, receivedCheck.Validate())
		assert.EqualValues(t, check, receivedCheck)
	})
	putReq := newRequest("POST", fmt.Sprintf("/checks/check1"), bytes.NewBuffer(updatedCheckJSON))
	putRes := processRequest(c, putReq)

	assert.Equal(t, http.StatusOK, putRes.Code)

	unauthReq := newRequest("POST", "/checks/check1", bytes.NewBuffer(updatedCheckJSON))
	unauthReq = requestWithNoAccess(unauthReq)

	unauthRes := processRequest(c, unauthReq)
	assert.Equal(t, http.StatusUnauthorized, unauthRes.Code)
}

func TestHttpApiCheckDelete(t *testing.T) {
	store := &mockstore.MockStore{}

	c := &ChecksController{
		Store: store,
	}

	check := types.FixtureCheckConfig("check1")
	store.On("GetCheckConfigByName", mock.Anything, "check1").Return(check, nil)
	store.On("DeleteCheckConfigByName", mock.Anything, "check1").Return(nil)
	deleteReq := newRequest("DELETE", fmt.Sprintf("/checks/check1"), nil)
	deleteRes := processRequest(c, deleteReq)

	assert.Equal(t, http.StatusOK, deleteRes.Code)
}

func TestHttpApiChecksDeleteUnauthorized(t *testing.T) {
	controller := ChecksController{}

	req := newRequest("DELETE", "/checks/check1", nil)
	req = requestWithNoAccess(req)

	res := processRequest(&controller, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
}*/

package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/sensu/sensu-go/testing/fixtures"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestHttpApiChecksGet(t *testing.T) {
	c := &ChecksController{
		Store: fixtures.NewFixtureStore(),
	}

	req, _ := http.NewRequest("GET", "/checks", nil)
	res := processRequest(c, req)

	assert.Equal(t, http.StatusOK, res.Code)

	body := res.Body.Bytes()

	checks := []*types.Check{}
	err := json.Unmarshal(body, &checks)

	assert.NoError(t, err)
	assert.Condition(t, func() bool { return len(checks) >= 1 })
}

func TestHttpApiCheckGet(t *testing.T) {
	c := &ChecksController{
		Store: fixtures.NewFixtureStore(),
	}

	notFoundReq, _ := http.NewRequest("GET", "/checks/somecheck", nil)
	notFoundRes := processRequest(c, notFoundReq)

	assert.Equal(t, http.StatusNotFound, notFoundRes.Code)

	foundReq, _ := http.NewRequest("GET", "/checks/check1", nil)
	foundRes := processRequest(c, foundReq)

	assert.Equal(t, http.StatusOK, foundRes.Code)

	body := foundRes.Body.Bytes()

	check := &types.Check{}
	err := json.Unmarshal(body, &check)

	assert.NoError(t, err)
	assert.NotNil(t, check.Name)
	assert.NotNil(t, check.Command)
	assert.NotEqual(t, check.Name, "")
	assert.NotEqual(t, check.Command, "")
}

func TestHttpApiCheckPut(t *testing.T) {
	c := &ChecksController{
		Store: fixtures.NewFixtureStore(),
	}

	checkName := "check1"

	updatedCheck := &types.Check{
		Name:     checkName,
		Interval: 120,
		Command:  "command2",
	}

	updatedCheckJSON, _ := json.Marshal(updatedCheck)

	putReq, _ := http.NewRequest("PUT", fmt.Sprintf("/checks/%s", checkName), bytes.NewBuffer(updatedCheckJSON))
	putRes := processRequest(c, putReq)

	assert.Equal(t, http.StatusOK, putRes.Code)

	getReq, _ := http.NewRequest("GET", fmt.Sprintf("/checks/%s", checkName), nil)
	getRes := processRequest(c, getReq)

	assert.Equal(t, http.StatusOK, getRes.Code)

	body := getRes.Body.String()

	assert.Equal(t, string(updatedCheckJSON[:]), body)
}

func TestHttpApiCheckPost(t *testing.T) {
	c := &ChecksController{
		Store: fixtures.NewFixtureStore(),
	}

	checkName := "newcheck1"

	updatedCheck := &types.Check{
		Name:     checkName,
		Interval: 60,
		Command:  "command2",
	}

	updatedCheckJSON, _ := json.Marshal(updatedCheck)

	putReq, _ := http.NewRequest("POST", fmt.Sprintf("/checks/%s", checkName), bytes.NewBuffer(updatedCheckJSON))
	putRes := processRequest(c, putReq)

	assert.Equal(t, http.StatusOK, putRes.Code)

	getReq, _ := http.NewRequest("GET", fmt.Sprintf("/checks/%s", checkName), nil)
	getRes := processRequest(c, getReq)

	assert.Equal(t, http.StatusOK, getRes.Code)

	body := getRes.Body.String()

	assert.Equal(t, string(updatedCheckJSON[:]), body)
}

func TestHttpApiCheckDelete(t *testing.T) {
	c := &ChecksController{
		Store: fixtures.NewFixtureStore(),
	}

	checkName := "check1"

	deleteReq, _ := http.NewRequest("DELETE", fmt.Sprintf("/checks/%s", checkName), nil)
	deleteRes := processRequest(c, deleteReq)

	assert.Equal(t, http.StatusOK, deleteRes.Code)

	getReq, _ := http.NewRequest("GET", fmt.Sprintf("/checks/%s", checkName), nil)
	getRes := processRequest(c, getReq)

	assert.Equal(t, http.StatusNotFound, getRes.Code)
}

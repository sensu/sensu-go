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

func TestHttpApiMutatorsGet(t *testing.T) {
	c := &MutatorsController{
		Store: fixtures.NewFixtureStore(),
	}

	req, _ := http.NewRequest("GET", "/mutators", nil)
	res := processRequest(c, req)

	assert.Equal(t, http.StatusOK, res.Code)

	body := res.Body.Bytes()

	mutators := []*types.Mutator{}
	err := json.Unmarshal(body, &mutators)

	assert.NoError(t, err)
	assert.Condition(t, func() bool { return len(mutators) >= 1 })
}

func TestHttpApiMutatorGet(t *testing.T) {
	c := &MutatorsController{
		Store: fixtures.NewFixtureStore(),
	}

	notFoundReq, _ := http.NewRequest("GET", "/mutators/somemutator", nil)
	notFoundRes := processRequest(c, notFoundReq)

	assert.Equal(t, http.StatusNotFound, notFoundRes.Code)

	foundReq, _ := http.NewRequest("GET", "/mutators/mutator1", nil)
	foundRes := processRequest(c, foundReq)

	assert.Equal(t, http.StatusOK, foundRes.Code)

	body := foundRes.Body.Bytes()

	mutator := &types.Mutator{}
	err := json.Unmarshal(body, &mutator)

	assert.NoError(t, err)
	assert.NotNil(t, mutator.Name)
	assert.NotNil(t, mutator.Command)
	assert.NotEqual(t, mutator.Name, "")
	assert.NotEqual(t, mutator.Command, "")
}

func TestHttpApiMutatorPut(t *testing.T) {
	c := &MutatorsController{
		Store: fixtures.NewFixtureStore(),
	}

	mutatorName := "newmutator1"

	updatedMutator := &types.Mutator{
		Name:    mutatorName,
		Command: "dog",
		Timeout: 50,
	}

	updatedMutatorJSON, _ := json.Marshal(updatedMutator)

	putReq, _ := http.NewRequest("PUT", fmt.Sprintf("/mutators/%s", mutatorName), bytes.NewBuffer(updatedMutatorJSON))
	putRes := processRequest(c, putReq)

	assert.Equal(t, http.StatusOK, putRes.Code)

	getReq, _ := http.NewRequest("GET", fmt.Sprintf("/mutators/%s", mutatorName), nil)
	getRes := processRequest(c, getReq)

	assert.Equal(t, http.StatusOK, getRes.Code)

	body := getRes.Body.String()

	assert.Equal(t, string(updatedMutatorJSON[:]), body)
}

func TestHttpApiMutatorPost(t *testing.T) {
	c := &MutatorsController{
		Store: fixtures.NewFixtureStore(),
	}

	mutatorName := "newmutator1"

	updatedMutator := &types.Mutator{
		Name:    mutatorName,
		Command: "cat",
		Timeout: 10,
	}

	updatedMutatorJSON, _ := json.Marshal(updatedMutator)

	putReq, _ := http.NewRequest("POST", fmt.Sprintf("/mutators/%s", mutatorName), bytes.NewBuffer(updatedMutatorJSON))
	putRes := processRequest(c, putReq)

	assert.Equal(t, http.StatusOK, putRes.Code)

	getReq, _ := http.NewRequest("GET", fmt.Sprintf("/mutators/%s", mutatorName), nil)
	getRes := processRequest(c, getReq)

	assert.Equal(t, http.StatusOK, getRes.Code)

	body := getRes.Body.String()

	assert.Equal(t, string(updatedMutatorJSON[:]), body)
}

func TestHttpApiMutatorDelete(t *testing.T) {
	c := &MutatorsController{
		Store: fixtures.NewFixtureStore(),
	}

	mutatorName := "mutator1"

	deleteReq, _ := http.NewRequest("DELETE", fmt.Sprintf("/mutators/%s", mutatorName), nil)
	deleteRes := processRequest(c, deleteReq)

	assert.Equal(t, http.StatusOK, deleteRes.Code)

	getReq, _ := http.NewRequest("GET", fmt.Sprintf("/mutators/%s", mutatorName), nil)
	getRes := processRequest(c, getReq)

	assert.Equal(t, http.StatusNotFound, getRes.Code)
}

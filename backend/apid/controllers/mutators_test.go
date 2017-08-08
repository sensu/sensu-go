package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHttpApiMutatorsGet(t *testing.T) {
	store := &mockstore.MockStore{}
	controller := MutatorsController{Store: store}

	mutators := []*types.Mutator{
		types.FixtureMutator("mutator1"),
		types.FixtureMutator("mutator2"),
	}

	store.On("GetMutators", mock.Anything).Return(mutators, nil)
	req := newRequest("GET", "/mutators", nil)
	res := processRequest(&controller, req)

	assert.Equal(t, http.StatusOK, res.Code)

	body := res.Body.Bytes()

	returnedMutators := []*types.Mutator{}
	err := json.Unmarshal(body, &returnedMutators)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(returnedMutators))
	for i, mutator := range returnedMutators {
		assert.EqualValues(t, mutators[i], mutator)
	}
}

func TestHttpApiMutatorsGetUnauthorized(t *testing.T) {
	controller := MutatorsController{}

	req := newRequest("GET", "/mutators", nil)
	req = requestWithNoAccess(req)

	res := processRequest(&controller, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestHttpApiMutatorGet(t *testing.T) {
	store := &mockstore.MockStore{}

	c := &MutatorsController{
		Store: store,
	}

	var nilMutator *types.Mutator
	store.On("GetMutatorByName", mock.Anything, "somemutator").Return(nilMutator, nil)
	notFoundReq := newRequest("GET", "/mutators/somemutator", nil)
	notFoundRes := processRequest(c, notFoundReq)

	assert.Equal(t, http.StatusNotFound, notFoundRes.Code)

	mutatorName := "mutator1"
	mutator := types.FixtureMutator(mutatorName)
	store.On("GetMutatorByName", mock.Anything, mutatorName).Return(mutator, nil)
	foundReq := newRequest("GET", fmt.Sprintf("/mutators/%s", mutatorName), nil)
	foundRes := processRequest(c, foundReq)

	assert.Equal(t, http.StatusOK, foundRes.Code)

	body := foundRes.Body.Bytes()

	returnedMutator := &types.Mutator{}
	err := json.Unmarshal(body, &returnedMutator)

	assert.NoError(t, err)
	assert.EqualValues(t, mutator, returnedMutator)
}

func TestHttpApiMutatorGetUnauthorized(t *testing.T) {
	store := &mockstore.MockStore{}
	controller := MutatorsController{Store: store}

	mutator := types.FixtureMutator("name")
	store.On("GetMutatorByName", mock.Anything, "name").Return(mutator, nil)

	req := newRequest("GET", "/mutators/name", nil)
	req = requestWithNoAccess(req)

	res := processRequest(&controller, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestHttpApiMutatorPut(t *testing.T) {
	store := &mockstore.MockStore{}

	c := &MutatorsController{
		Store: store,
	}

	mutatorName := "mutator1"
	mutator := types.FixtureMutator(mutatorName)

	updatedMutatorJSON, _ := json.Marshal(mutator)

	store.On("UpdateMutator", mock.AnythingOfType("*types.Mutator")).Return(nil).Run(func(args mock.Arguments) {
		receivedMutator := args.Get(0).(*types.Mutator)
		assert.EqualValues(t, mutator, receivedMutator)
	})
	putReq := newRequest("PUT", fmt.Sprintf("/mutators/%s", mutatorName), bytes.NewBuffer(updatedMutatorJSON))
	putRes := processRequest(c, putReq)

	assert.Equal(t, http.StatusOK, putRes.Code)
}

func TestHttpApiMutatorPost(t *testing.T) {
	store := &mockstore.MockStore{}

	c := &MutatorsController{
		Store: store,
	}

	mutatorName := "newmutator1"
	mutator := types.FixtureMutator(mutatorName)

	updatedMutatorJSON, _ := json.Marshal(mutator)

	store.On("UpdateMutator", mock.AnythingOfType("*types.Mutator")).Return(nil).Run(func(args mock.Arguments) {
		receivedMutator := args.Get(0).(*types.Mutator)
		assert.EqualValues(t, mutator, receivedMutator)
	})
	putReq := newRequest("POST", fmt.Sprintf("/mutators/%s", mutatorName), bytes.NewBuffer(updatedMutatorJSON))
	putRes := processRequest(c, putReq)

	assert.Equal(t, http.StatusOK, putRes.Code)

	unauthReq := newRequest("POST", "/mutators/"+mutatorName, bytes.NewBuffer(updatedMutatorJSON))
	unauthReq = requestWithNoAccess(unauthReq)

	unauthRes := processRequest(c, unauthReq)
	assert.Equal(t, http.StatusUnauthorized, unauthRes.Code)
}

func TestHttpApiMutatorDelete(t *testing.T) {
	store := &mockstore.MockStore{}

	c := &MutatorsController{
		Store: store,
	}

	mutatorName := "mutator1"
	mutator := types.FixtureMutator(mutatorName)
	store.On("GetMutatorByName", mock.Anything, mutatorName).Return(mutator, nil)
	store.On("DeleteMutatorByName", mock.Anything, mutatorName).Return(nil)
	deleteReq := newRequest("DELETE", fmt.Sprintf("/mutators/%s", mutatorName), nil)
	deleteRes := processRequest(c, deleteReq)

	assert.Equal(t, http.StatusOK, deleteRes.Code)
}

func TestHttpApiMutatorDeleteUnauthorized(t *testing.T) {
	controller := MutatorsController{}

	req := newRequest("DELETE", "/mutators/test", nil)
	req = requestWithNoAccess(req)

	res := processRequest(&controller, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
}

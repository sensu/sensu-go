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

	c := &MutatorsController{
		Store: store,
	}

	mutators := []*types.Mutator{
		types.FixtureMutator("mutator1"),
		types.FixtureMutator("mutator2"),
	}

	store.On("GetMutators", "default").Return(mutators, nil)
	req, _ := http.NewRequest("GET", "/mutators", nil)
	res := processRequest(c, req)

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

func TestHttpApiMutatorGet(t *testing.T) {
	store := &mockstore.MockStore{}

	c := &MutatorsController{
		Store: store,
	}

	var nilMutator *types.Mutator
	store.On("GetMutatorByName", "default", "somemutator").Return(nilMutator, nil)
	notFoundReq, _ := http.NewRequest("GET", "/mutators/somemutator", nil)
	notFoundRes := processRequest(c, notFoundReq)

	assert.Equal(t, http.StatusNotFound, notFoundRes.Code)

	mutatorName := "mutator1"
	mutator := types.FixtureMutator(mutatorName)
	store.On("GetMutatorByName", "default", mutatorName).Return(mutator, nil)
	foundReq, _ := http.NewRequest("GET", fmt.Sprintf("/mutators/%s", mutatorName), nil)
	foundRes := processRequest(c, foundReq)

	assert.Equal(t, http.StatusOK, foundRes.Code)

	body := foundRes.Body.Bytes()

	returnedMutator := &types.Mutator{}
	err := json.Unmarshal(body, &returnedMutator)

	assert.NoError(t, err)
	assert.EqualValues(t, mutator, returnedMutator)
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
	putReq, _ := http.NewRequest("PUT", fmt.Sprintf("/mutators/%s", mutatorName), bytes.NewBuffer(updatedMutatorJSON))
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
	putReq, _ := http.NewRequest("POST", fmt.Sprintf("/mutators/%s", mutatorName), bytes.NewBuffer(updatedMutatorJSON))
	putRes := processRequest(c, putReq)

	assert.Equal(t, http.StatusOK, putRes.Code)
}

func TestHttpApiMutatorDelete(t *testing.T) {
	store := &mockstore.MockStore{}

	c := &MutatorsController{
		Store: store,
	}

	mutatorName := "mutator1"
	mutator := types.FixtureMutator(mutatorName)
	store.On("GetMutatorByName", "default", mutatorName).Return(mutator, nil)
	store.On("DeleteMutatorByName", "default", mutatorName).Return(nil)
	deleteReq, _ := http.NewRequest("DELETE", fmt.Sprintf("/mutators/%s", mutatorName), nil)
	deleteRes := processRequest(c, deleteReq)

	assert.Equal(t, http.StatusOK, deleteRes.Code)
}

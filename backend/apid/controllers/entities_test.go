package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHttpApiEntitiesGet(t *testing.T) {
	store := &mockstore.MockStore{}

	c := &EntitiesController{
		Store: store,
	}

	entities := []*types.Entity{
		types.FixtureEntity("entity1"),
		types.FixtureEntity("entity2"),
	}
	store.On("GetEntities", mock.Anything).Return(entities, nil)
	req, _ := http.NewRequest("GET", "/entities", nil)
	req = requestWithDefaultContext(req)
	res := processRequest(c, req)

	assert.Equal(t, http.StatusOK, res.Code)

	body := res.Body.Bytes()

	returnedEntities := []*types.Entity{}
	err := json.Unmarshal(body, &returnedEntities)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(returnedEntities))
	for i, entity := range returnedEntities {
		assert.EqualValues(t, entities[i], entity)
	}
}

func TestHttpApiEntityGet(t *testing.T) {
	store := &mockstore.MockStore{}

	c := &EntitiesController{
		Store: store,
	}

	var nilEntity *types.Entity
	store.On("GetEntityByID", mock.Anything, "someentity").Return(nilEntity, nil)
	notFoundReq, _ := http.NewRequest("GET", "/entities/someentity", nil)
	notFoundReq = requestWithDefaultContext(notFoundReq)
	notFoundRes := processRequest(c, notFoundReq)

	assert.Equal(t, http.StatusNotFound, notFoundRes.Code)

	entity1 := types.FixtureEntity("entity1")
	store.On("GetEntityByID", mock.Anything, "entity1").Return(entity1, nil)
	foundReq, _ := http.NewRequest("GET", "/entities/entity1", nil)
	foundReq = requestWithDefaultContext(foundReq)
	foundRes := processRequest(c, foundReq)

	assert.Equal(t, http.StatusOK, foundRes.Code)

	body := foundRes.Body.Bytes()

	returnedEntity := &types.Entity{}
	err := json.Unmarshal(body, &returnedEntity)

	assert.NoError(t, err)
	assert.EqualValues(t, entity1, returnedEntity)
}

func TestHttpApiEntityDelete(t *testing.T) {
	assert := assert.New(t)

	store := &mockstore.MockStore{}

	c := &EntitiesController{
		Store: store,
	}

	entity := types.FixtureEntity("entity1")
	store.On("GetEntityByID", mock.Anything, "entity1").Return(entity, nil)
	store.On("DeleteEntityByID", mock.Anything, "entity1").Return(nil)
	deleteReq := newRequest("DELETE", fmt.Sprintf("/entities/entity1"), nil)
	deleteRes := processRequest(c, deleteReq)

	store.AssertCalled(t, "GetEntityByID", mock.Anything, "entity1")
	store.AssertCalled(t, "DeleteEntityByID", mock.Anything, "entity1")

	assert.Equal(http.StatusOK, deleteRes.Code)
}

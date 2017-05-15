package controllers

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
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
	store.On("GetEntities").Return(entities, nil)
	req, _ := http.NewRequest("GET", "/entities", nil)
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

func TestHttpApiEntitiesGetEmpty(t *testing.T) {
	store := &mockstore.MockStore{}

	c := &EntitiesController{
		Store: store,
	}

	var entities []*types.Entity
	store.On("GetEntities").Return(entities, nil)
	req, _ := http.NewRequest("GET", "/entities", nil)
	res := processRequest(c, req)

	assert.Equal(t, http.StatusOK, res.Code)

	body := res.Body.String()
	assert.Equal(t, "[]", body)
}

func TestHttpApiEntityGet(t *testing.T) {
	store := &mockstore.MockStore{}

	c := &EntitiesController{
		Store: store,
	}

	var nilEntity *types.Entity
	store.On("GetEntityByID", "someentity").Return(nilEntity, nil)
	notFoundReq, _ := http.NewRequest("GET", "/entities/someentity", nil)
	notFoundRes := processRequest(c, notFoundReq)

	assert.Equal(t, http.StatusNotFound, notFoundRes.Code)

	entity1 := types.FixtureEntity("entity1")
	store.On("GetEntityByID", "entity1").Return(entity1, nil)
	foundReq, _ := http.NewRequest("GET", "/entities/entity1", nil)
	foundRes := processRequest(c, foundReq)

	assert.Equal(t, http.StatusOK, foundRes.Code)

	body := foundRes.Body.Bytes()

	returnedEntity := &types.Entity{}
	err := json.Unmarshal(body, &returnedEntity)

	assert.NoError(t, err)
	assert.EqualValues(t, entity1, returnedEntity)
}

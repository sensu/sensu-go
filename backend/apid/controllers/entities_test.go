package controllers

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/sensu/sensu-go/testing/fixtures"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestHttpApiEntitiesGet(t *testing.T) {
	c := &EntitiesController{
		Store: fixtures.NewFixtureStore(),
	}

	req, _ := http.NewRequest("GET", "/entities", nil)
	res := processRequest(c, req)

	assert.Equal(t, http.StatusOK, res.Code)

	body := res.Body.Bytes()

	entities := []*types.Entity{}
	err := json.Unmarshal(body, &entities)

	assert.NoError(t, err)
	assert.Condition(t, func() bool { return len(entities) >= 1 })
}

func TestHttpApiEntityGet(t *testing.T) {
	c := &EntitiesController{
		Store: fixtures.NewFixtureStore(),
	}

	notFoundReq, _ := http.NewRequest("GET", "/entities/someentity", nil)
	notFoundRes := processRequest(c, notFoundReq)

	assert.Equal(t, http.StatusNotFound, notFoundRes.Code)

	foundReq, _ := http.NewRequest("GET", "/entities/entity1", nil)
	foundRes := processRequest(c, foundReq)

	assert.Equal(t, http.StatusOK, foundRes.Code)

	body := foundRes.Body.Bytes()

	entity := &types.Entity{}
	err := json.Unmarshal(body, &entity)

	assert.NoError(t, err)
	assert.NotNil(t, entity.ID)
	assert.NotNil(t, entity.Class)
	assert.NotEqual(t, entity.ID, "")
	assert.Equal(t, entity.Class, "")
}

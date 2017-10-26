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

func TestDeleteFilter(t *testing.T) {
	store := &mockstore.MockStore{}

	c := &FiltersController{
		Store: store,
	}

	store.On("DeleteEventFilterByName", mock.Anything, "filter1").Return(nil)
	deleteReq := newRequest("DELETE", fmt.Sprintf("/filters/%s", "filter1"), nil)
	deleteRes := processRequest(c, deleteReq)
	assert.Equal(t, http.StatusOK, deleteRes.Code)

	store.On("DeleteEventFilterByName", mock.Anything, "filter2").Return(fmt.Errorf("error"))
	deleteReq = newRequest("DELETE", fmt.Sprintf("/filters/%s", "filter2"), nil)
	deleteRes = processRequest(c, deleteReq)
	assert.Equal(t, http.StatusInternalServerError, deleteRes.Code)
}

func TestGetFilter(t *testing.T) {
	store := &mockstore.MockStore{}

	c := &FiltersController{
		Store: store,
	}

	var nilFilter *types.EventFilter
	store.On("GetEventFilterByName", mock.Anything, "foo").Return(nilFilter, nil)
	notFoundReq := newRequest("GET", "/filters/foo", nil)
	notFoundRes := processRequest(c, notFoundReq)
	assert.Equal(t, http.StatusNotFound, notFoundRes.Code)

	filter := types.FixtureEventFilter("filter1")
	store.On("GetEventFilterByName", mock.Anything, "filter1").Return(filter, nil)
	foundReq := newRequest("GET", "/filters/filter1", nil)
	foundRes := processRequest(c, foundReq)
	assert.Equal(t, http.StatusOK, foundRes.Code)

	body := foundRes.Body.Bytes()

	result := &types.EventFilter{}
	err := json.Unmarshal(body, &result)

	assert.NoError(t, err)
	assert.EqualValues(t, filter, result)
}

func TestGetFilters(t *testing.T) {
	store := &mockstore.MockStore{}

	c := &FiltersController{
		Store: store,
	}

	filters := []*types.EventFilter{
		types.FixtureEventFilter("filter1"),
		types.FixtureEventFilter("filter2"),
	}
	store.On("GetEventFilters", mock.Anything).Return(filters, nil)

	req := newRequest("GET", "/filters", nil)
	res := processRequest(c, req)
	assert.Equal(t, http.StatusOK, res.Code)

	body := res.Body.Bytes()

	result := []*types.EventFilter{}
	err := json.Unmarshal(body, &result)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	for i, filter := range result {
		assert.EqualValues(t, filters[i], filter)
	}
}

func TestPostPutFilter(t *testing.T) {
	store := &mockstore.MockStore{}

	c := &FiltersController{
		Store: store,
	}

	name := "filter1"
	filter := types.FixtureEventFilter(name)
	filterBytes, _ := json.Marshal(filter)

	// Successful
	store.On("UpdateEventFilter", filter).Return(nil).Run(func(args mock.Arguments) {
		receivedFilter := args.Get(0).(*types.EventFilter)
		assert.EqualValues(t, filter, receivedFilter)
	})
	putReq := newRequest("POST", fmt.Sprintf("/filters/%s", name), bytes.NewBuffer(filterBytes))
	putRes := processRequest(c, putReq)
	assert.Equal(t, http.StatusOK, putRes.Code)

	// Bad Request-URI
	putReq = newRequest("POST", fmt.Sprintf("/filters/%s", "foobar"), bytes.NewBuffer(filterBytes))
	putRes = processRequest(c, putReq)
	assert.Equal(t, http.StatusBadRequest, putRes.Code)

	// Error from the store
	name = "filter2"
	filter = types.FixtureEventFilter(name)
	filterBytes, _ = json.Marshal(filter)
	store.On("UpdateEventFilter", mock.AnythingOfType("*types.EventFilter")).Return(fmt.Errorf("error"))
	putReq = newRequest("POST", fmt.Sprintf("/filters/%s", name), bytes.NewBuffer(filterBytes))
	putRes = processRequest(c, putReq)
	assert.Equal(t, http.StatusInternalServerError, putRes.Code)

	// Unauthorized
	unauthReq := newRequest("POST", "/filters/"+name, bytes.NewBuffer(filterBytes))
	unauthReq = requestWithNoAccess(unauthReq)
	unauthRes := processRequest(c, unauthReq)
	assert.Equal(t, http.StatusUnauthorized, unauthRes.Code)
}

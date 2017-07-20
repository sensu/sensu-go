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

func TestHttpAPIHandlersGet(t *testing.T) {
	store := &mockstore.MockStore{}

	c := &HandlersController{
		Store: store,
	}

	handlers := []*types.Handler{
		types.FixtureHandler("handler1"),
		types.FixtureHandler("handler2"),
	}
	store.On("GetHandlers", mock.Anything).Return(handlers, nil)
	req := newRequest("GET", "/handlers", nil)
	res := processRequest(c, req)

	assert.Equal(t, http.StatusOK, res.Code)

	body := res.Body.Bytes()

	receivedHandlers := []*types.Handler{}
	err := json.Unmarshal(body, &receivedHandlers)

	assert.NoError(t, err)
	assert.Equal(t, 2, len(receivedHandlers))
	for i, handler := range receivedHandlers {
		assert.EqualValues(t, handlers[i], handler)
	}
}

func TestHttpAPIHandlersGetUnauthorized(t *testing.T) {
	controller := HandlersController{}

	req := newRequest("GET", "/handlers", nil)
	req = requestWithNoAccess(req)

	res := processRequest(&controller, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestHttpAPIHandlerGet(t *testing.T) {
	store := &mockstore.MockStore{}

	c := &HandlersController{
		Store: store,
	}

	var nilHandler *types.Handler
	store.On("GetHandlerByName", mock.Anything, "somehandler").Return(nilHandler, nil)
	notFoundReq := newRequest("GET", "/handlers/somehandler", nil)
	notFoundRes := processRequest(c, notFoundReq)

	assert.Equal(t, http.StatusNotFound, notFoundRes.Code)

	handler := types.FixtureHandler("handler1")
	store.On("GetHandlerByName", mock.Anything, "handler1").Return(handler, nil)
	foundReq := newRequest("GET", "/handlers/handler1", nil)
	foundRes := processRequest(c, foundReq)

	assert.Equal(t, http.StatusOK, foundRes.Code)

	body := foundRes.Body.Bytes()

	receivedHandler := &types.Handler{}
	err := json.Unmarshal(body, &receivedHandler)

	assert.NoError(t, err)
	assert.EqualValues(t, handler, receivedHandler)
}

func TestHttpAPIHandlerGetUnauthorized(t *testing.T) {
	controller := HandlersController{}

	req := newRequest("GET", "/handlers/meow", nil)
	req = requestWithNoAccess(req)

	res := processRequest(&controller, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestHttpAPIHandlerPut(t *testing.T) {
	store := &mockstore.MockStore{}

	c := &HandlersController{
		Store: store,
	}

	handler := types.FixtureHandler("handler1")
	updatedHandlerJSON, _ := json.Marshal(handler)

	store.On("UpdateHandler", mock.AnythingOfType("*types.Handler")).Return(nil).Run(func(args mock.Arguments) {
		receivedHandler := args.Get(0).(*types.Handler)
		assert.EqualValues(t, handler, receivedHandler)
	})
	putReq := newRequest("PUT", fmt.Sprintf("/handlers/handler1"), bytes.NewBuffer(updatedHandlerJSON))
	putRes := processRequest(c, putReq)

	assert.Equal(t, http.StatusOK, putRes.Code)

	unauthReq := newRequest("PUT", "/handlers/"+handler.Name, nil)
	unauthReq = requestWithNoAccess(unauthReq)

	unauthRes := processRequest(c, unauthReq)
	assert.Equal(t, http.StatusUnauthorized, unauthRes.Code)
}

func TestHttpAPIHandlerPost(t *testing.T) {
	store := &mockstore.MockStore{}

	c := &HandlersController{
		Store: store,
	}

	handlerName := "newhandler1"

	handler := types.FixtureHandler(handlerName)
	updatedHandlerJSON, _ := json.Marshal(handler)

	store.On("UpdateHandler", mock.AnythingOfType("*types.Handler")).Return(nil).Run(func(args mock.Arguments) {
		receivedHandler := args.Get(0).(*types.Handler)
		assert.EqualValues(t, handler, receivedHandler)
	})

	putReq := newRequest("POST", fmt.Sprintf("/handlers/%s", handlerName), bytes.NewBuffer(updatedHandlerJSON))
	putRes := processRequest(c, putReq)

	assert.Equal(t, http.StatusOK, putRes.Code)

	unauthReq := newRequest("POST", "/handlers/"+handler.Name, nil)
	unauthReq = requestWithNoAccess(unauthReq)

	unauthRes := processRequest(c, unauthReq)
	assert.Equal(t, http.StatusUnauthorized, unauthRes.Code)
}

func TestHttpAPIHandlerDelete(t *testing.T) {
	store := &mockstore.MockStore{}

	c := &HandlersController{
		Store: store,
	}

	handlerName := "handler1"

	handler := types.FixtureHandler(handlerName)

	store.On("GetHandlerByName", mock.Anything, handlerName).Return(handler, nil)
	store.On("DeleteHandlerByName", mock.Anything, handlerName).Return(nil)
	deleteReq := newRequest("DELETE", fmt.Sprintf("/handlers/%s", handlerName), nil)
	deleteRes := processRequest(c, deleteReq)

	assert.Equal(t, http.StatusOK, deleteRes.Code)
}

func TestHttpAPIHandlerDeleteUnauthorized(t *testing.T) {
	controller := HandlersController{}

	deleteReq := newRequest("DELETE", "/handlers/meow", nil)
	deleteReq = requestWithNoAccess(deleteReq)

	deleteRes := processRequest(&controller, deleteReq)
	assert.Equal(t, http.StatusUnauthorized, deleteRes.Code)
}

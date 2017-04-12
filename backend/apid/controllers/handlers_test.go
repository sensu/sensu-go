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

func TestHttpAPIHandlersGet(t *testing.T) {
	c := &HandlersController{
		Store: fixtures.NewFixtureStore(),
	}

	req, _ := http.NewRequest("GET", "/handlers", nil)
	res := processRequest(c, req)

	assert.Equal(t, http.StatusOK, res.Code)

	body := res.Body.Bytes()

	handlers := []*types.Handler{}
	err := json.Unmarshal(body, &handlers)

	assert.NoError(t, err)
	assert.Condition(t, func() bool { return len(handlers) >= 1 })
}

func TestHttpAPIHandlerGet(t *testing.T) {
	c := &HandlersController{
		Store: fixtures.NewFixtureStore(),
	}

	notFoundReq, _ := http.NewRequest("GET", "/handlers/somehandler", nil)
	notFoundRes := processRequest(c, notFoundReq)

	assert.Equal(t, http.StatusNotFound, notFoundRes.Code)

	foundReq, _ := http.NewRequest("GET", "/handlers/handler1", nil)
	foundRes := processRequest(c, foundReq)

	assert.Equal(t, http.StatusOK, foundRes.Code)

	body := foundRes.Body.Bytes()

	handler := &types.Handler{}
	err := json.Unmarshal(body, &handler)

	assert.NoError(t, err)
	assert.NotNil(t, handler.Name)
	assert.NotNil(t, handler.Type)
	assert.NotEqual(t, handler.Name, "")
	assert.NotEqual(t, handler.Type, "")
}

func TestHttpAPIHandlerPut(t *testing.T) {
	c := &HandlersController{
		Store: fixtures.NewFixtureStore(),
	}

	handlerName := "handler1"

	updatedHandler := &types.Handler{
		Name:    handlerName,
		Type:    "pipe",
		Mutator: "mutator2",
		Command: "cat",
		Timeout: 10,
	}

	updatedHandlerJSON, _ := json.Marshal(updatedHandler)

	putReq, _ := http.NewRequest("PUT", fmt.Sprintf("/handlers/%s", handlerName), bytes.NewBuffer(updatedHandlerJSON))
	putRes := processRequest(c, putReq)

	assert.Equal(t, http.StatusOK, putRes.Code)

	getReq, _ := http.NewRequest("GET", fmt.Sprintf("/handlers/%s", handlerName), nil)
	getRes := processRequest(c, getReq)

	assert.Equal(t, http.StatusOK, getRes.Code)

	body := getRes.Body.String()

	assert.Equal(t, string(updatedHandlerJSON[:]), body)
}

func TestHttpAPIHandlerPost(t *testing.T) {
	c := &HandlersController{
		Store: fixtures.NewFixtureStore(),
	}

	handlerName := "newhandler1"

	updatedHandler := &types.Handler{
		Name:    handlerName,
		Type:    "pipe",
		Mutator: "mutator2",
		Command: "cat",
		Timeout: 10,
	}

	updatedHandlerJSON, _ := json.Marshal(updatedHandler)

	putReq, _ := http.NewRequest("POST", fmt.Sprintf("/handlers/%s", handlerName), bytes.NewBuffer(updatedHandlerJSON))
	putRes := processRequest(c, putReq)

	assert.Equal(t, http.StatusOK, putRes.Code)

	getReq, _ := http.NewRequest("GET", fmt.Sprintf("/handlers/%s", handlerName), nil)
	getRes := processRequest(c, getReq)

	assert.Equal(t, http.StatusOK, getRes.Code)

	body := getRes.Body.String()

	assert.Equal(t, string(updatedHandlerJSON[:]), body)
}

func TestHttpAPIHandlerDelete(t *testing.T) {
	c := &HandlersController{
		Store: fixtures.NewFixtureStore(),
	}

	handlerName := "handler1"

	deleteReq, _ := http.NewRequest("DELETE", fmt.Sprintf("/handlers/%s", handlerName), nil)
	deleteRes := processRequest(c, deleteReq)

	assert.Equal(t, http.StatusOK, deleteRes.Code)

	getReq, _ := http.NewRequest("GET", fmt.Sprintf("/handlers/%s", handlerName), nil)
	getRes := processRequest(c, getReq)

	assert.Equal(t, http.StatusNotFound, getRes.Code)
}

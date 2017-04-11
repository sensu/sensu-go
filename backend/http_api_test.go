package backend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sensu/sensu-go/testing/fixtures"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func getAPI() *API {
	store := fixtures.NewFixtureStore()

	api := &API{
		Store: store,
	}

	return api
}

func processRequest(req *http.Request, api *API) *httptest.ResponseRecorder {
	if api == nil {
		api = getAPI()
	}

	router := httpRouter(api)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	return res
}

func TestHttpApiHandlersHandlerGet(t *testing.T) {
	req, _ := http.NewRequest("GET", "/handlers", nil)
	res := processRequest(req, nil)

	assert.Equal(t, http.StatusOK, res.Code)

	body := res.Body.Bytes()

	handlers := []*types.Handler{}
	err := json.Unmarshal(body, &handlers)

	assert.NoError(t, err)
	assert.Condition(t, func() bool { return len(handlers) >= 1 })
}

func TestHttpApiHandlerHandlerGet(t *testing.T) {
	api := getAPI()

	notFoundReq, _ := http.NewRequest("GET", "/handlers/somehandler", nil)
	notFoundRes := processRequest(notFoundReq, api)

	assert.Equal(t, http.StatusNotFound, notFoundRes.Code)

	foundReq, _ := http.NewRequest("GET", "/handlers/handler1", nil)
	foundRes := processRequest(foundReq, api)

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

func TestHttpApiHandlerHandlerPut(t *testing.T) {
	api := getAPI()

	handlerName := "handler1"

	updatedHandler := &types.Handler{
		Name:    handlerName,
		Type:    "pipe",
		Mutator: "mutator2",
		Pipe: types.HandlerPipe{
			Command: "cat",
			Timeout: 10,
		},
	}

	updatedHandlerJSON, _ := json.Marshal(updatedHandler)

	putReq, _ := http.NewRequest("PUT", fmt.Sprintf("/handlers/%s", handlerName), bytes.NewBuffer(updatedHandlerJSON))
	putRes := processRequest(putReq, api)

	assert.Equal(t, http.StatusOK, putRes.Code)

	getReq, _ := http.NewRequest("GET", fmt.Sprintf("/handlers/%s", handlerName), nil)
	getRes := processRequest(getReq, api)

	assert.Equal(t, http.StatusOK, getRes.Code)

	body := getRes.Body.String()

	assert.Equal(t, string(updatedHandlerJSON[:]), body)
}

func TestHttpApiHandlerHandlerPost(t *testing.T) {
	api := getAPI()

	handlerName := "newhandler1"

	updatedHandler := &types.Handler{
		Name:    handlerName,
		Type:    "pipe",
		Mutator: "mutator2",
		Pipe: types.HandlerPipe{
			Command: "cat",
			Timeout: 10,
		},
	}

	updatedHandlerJSON, _ := json.Marshal(updatedHandler)

	putReq, _ := http.NewRequest("POST", fmt.Sprintf("/handlers/%s", handlerName), bytes.NewBuffer(updatedHandlerJSON))
	putRes := processRequest(putReq, api)

	assert.Equal(t, http.StatusOK, putRes.Code)

	getReq, _ := http.NewRequest("GET", fmt.Sprintf("/handlers/%s", handlerName), nil)
	getRes := processRequest(getReq, api)

	assert.Equal(t, http.StatusOK, getRes.Code)

	body := getRes.Body.String()

	assert.Equal(t, string(updatedHandlerJSON[:]), body)
}

func TestHttpApiHandlerHandlerDelete(t *testing.T) {
	api := getAPI()

	handlerName := "handler1"

	deleteReq, _ := http.NewRequest("DELETE", fmt.Sprintf("/handlers/%s", handlerName), nil)
	deleteRes := processRequest(deleteReq, api)

	assert.Equal(t, http.StatusOK, deleteRes.Code)

	getReq, _ := http.NewRequest("GET", fmt.Sprintf("/handlers/%s", handlerName), nil)
	getRes := processRequest(getReq, api)

	assert.Equal(t, http.StatusNotFound, getRes.Code)
}

func TestHttpApiMutatorsHandlerGet(t *testing.T) {
	req, _ := http.NewRequest("GET", "/mutators", nil)
	res := processRequest(req, nil)

	assert.Equal(t, http.StatusOK, res.Code)

	body := res.Body.Bytes()

	mutators := []*types.Mutator{}
	err := json.Unmarshal(body, &mutators)

	assert.NoError(t, err)
	assert.Condition(t, func() bool { return len(mutators) >= 1 })
}

func TestHttpApiMutatorHandlerGet(t *testing.T) {
	api := getAPI()

	notFoundReq, _ := http.NewRequest("GET", "/mutators/somemutator", nil)
	notFoundRes := processRequest(notFoundReq, api)

	assert.Equal(t, http.StatusNotFound, notFoundRes.Code)

	foundReq, _ := http.NewRequest("GET", "/mutators/mutator1", nil)
	foundRes := processRequest(foundReq, api)

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

func TestHttpApiMutatorHandlerPut(t *testing.T) {
	api := getAPI()

	mutatorName := "newmutator1"

	updatedMutator := &types.Mutator{
		Name:    mutatorName,
		Command: "dog",
		Timeout: 50,
	}

	updatedMutatorJSON, _ := json.Marshal(updatedMutator)

	putReq, _ := http.NewRequest("PUT", fmt.Sprintf("/mutators/%s", mutatorName), bytes.NewBuffer(updatedMutatorJSON))
	putRes := processRequest(putReq, api)

	assert.Equal(t, http.StatusOK, putRes.Code)

	getReq, _ := http.NewRequest("GET", fmt.Sprintf("/mutators/%s", mutatorName), nil)
	getRes := processRequest(getReq, api)

	assert.Equal(t, http.StatusOK, getRes.Code)

	body := getRes.Body.String()

	assert.Equal(t, string(updatedMutatorJSON[:]), body)
}

func TestHttpApiMutatorHandlerPost(t *testing.T) {
	api := getAPI()

	mutatorName := "newmutator1"

	updatedMutator := &types.Mutator{
		Name:    mutatorName,
		Command: "cat",
		Timeout: 10,
	}

	updatedMutatorJSON, _ := json.Marshal(updatedMutator)

	putReq, _ := http.NewRequest("POST", fmt.Sprintf("/mutators/%s", mutatorName), bytes.NewBuffer(updatedMutatorJSON))
	putRes := processRequest(putReq, api)

	assert.Equal(t, http.StatusOK, putRes.Code)

	getReq, _ := http.NewRequest("GET", fmt.Sprintf("/mutators/%s", mutatorName), nil)
	getRes := processRequest(getReq, api)

	assert.Equal(t, http.StatusOK, getRes.Code)

	body := getRes.Body.String()

	assert.Equal(t, string(updatedMutatorJSON[:]), body)
}

func TestHttpApiMutatorHandlerDelete(t *testing.T) {
	api := getAPI()

	mutatorName := "mutator1"

	deleteReq, _ := http.NewRequest("DELETE", fmt.Sprintf("/mutators/%s", mutatorName), nil)
	deleteRes := processRequest(deleteReq, api)

	assert.Equal(t, http.StatusOK, deleteRes.Code)

	getReq, _ := http.NewRequest("GET", fmt.Sprintf("/mutators/%s", mutatorName), nil)
	getRes := processRequest(getReq, api)

	assert.Equal(t, http.StatusNotFound, getRes.Code)
}

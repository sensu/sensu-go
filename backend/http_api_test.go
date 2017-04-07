package backend

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sensu/sensu-go/testing/fixtures"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func getApi() *API {
	store := fixtures.NewFixtureStore()

	api := &API{
		Store: store,
	}

	return api
}

func processRequest(req *http.Request, api *API) *httptest.ResponseRecorder {
	if api == nil {
		api = getApi()
	}

	router := httpRouter(api)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	return res
}

func TestHttpApiHandlersHandler(t *testing.T) {
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
	api := getApi()

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
	api := getApi()

	updatedHandler := &types.Handler{
		Name:    "handler1",
		Type:    "pipe",
		Mutator: "mutator2",
		Pipe: types.HandlerPipe{
			Command: "cat",
			Timeout: 10,
		},
	}

	updatedHandlerJson, _ := json.Marshal(updatedHandler)

	putReq, _ := http.NewRequest("PUT", "/handlers/handler1", bytes.NewBuffer(updatedHandlerJson))
	putRes := processRequest(putReq, api)

	assert.Equal(t, http.StatusOK, putRes.Code)

	getReq, _ := http.NewRequest("GET", "/handlers/handler1", nil)
	getRes := processRequest(getReq, api)

	assert.Equal(t, http.StatusOK, getRes.Code)

	body := getRes.Body.String()

	assert.Equal(t, string(updatedHandlerJson[:]), body)
}

func TestHttpApiHandlerHandlerPost(t *testing.T) {
	api := getApi()

	updatedHandler := &types.Handler{
		Name:    "newhandler1",
		Type:    "pipe",
		Mutator: "mutator2",
		Pipe: types.HandlerPipe{
			Command: "cat",
			Timeout: 10,
		},
	}

	updatedHandlerJson, _ := json.Marshal(updatedHandler)

	putReq, _ := http.NewRequest("POST", "/handlers/newhandler1", bytes.NewBuffer(updatedHandlerJson))
	putRes := processRequest(putReq, api)

	assert.Equal(t, http.StatusOK, putRes.Code)

	getReq, _ := http.NewRequest("GET", "/handlers/newhandler1", nil)
	getRes := processRequest(getReq, api)

	assert.Equal(t, http.StatusOK, getRes.Code)

	body := getRes.Body.String()

	assert.Equal(t, string(updatedHandlerJson[:]), body)
}

func TestHttpApiHandlerHandlerDelete(t *testing.T) {
	api := getApi()

	deleteReq, _ := http.NewRequest("DELETE", "/handlers/handler1", nil)
	deleteRes := processRequest(deleteReq, api)

	assert.Equal(t, http.StatusOK, deleteRes.Code)

	getReq, _ := http.NewRequest("GET", "/handlers/handler1", nil)
	getRes := processRequest(getReq, api)

	assert.Equal(t, http.StatusNotFound, getRes.Code)
}

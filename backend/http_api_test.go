package backend

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sensu/sensu-go/testing/fixtures"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func processRequest(req *http.Request) *httptest.ResponseRecorder {
	store := fixtures.NewFixtureStore()

	api := &API{
		Store: store,
	}

	router := httpRouter(api)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	return res
}

func TestHttpApiHandlersHandler(t *testing.T) {
	req, _ := http.NewRequest("GET", "/handlers", nil)
	res := processRequest(req)

	assert.Equal(t, http.StatusOK, res.Code)

	body := res.Body.Bytes()

	handlers := []*types.Handler{}
	err := json.Unmarshal(body, &handlers)

	assert.NoError(t, err)
	assert.Condition(t, func() bool { return len(handlers) >= 1 })
}

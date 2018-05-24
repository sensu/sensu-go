package routers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/graphql-go/graphql/testutil"
	"github.com/sensu/sensu-go/testing/mockbus"
	"github.com/sensu/sensu-go/testing/mockqueue"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/mock"
)

func setupGraphQLRouter() *GraphQLRouter {
	store := &mockstore.MockStore{}
	queue := &mockqueue.MockQueue{}
	bus := &mockbus.MockBus{}

	getter := &mockqueue.Getter{}
	getter.On("GetQueue", mock.Anything).Return(queue)

	router := NewGraphQLRouter(store, bus, getter)
	return router
}

func setupRequest(method string, path string, payload interface{}) (*http.Request, error) {
	reqPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, "/graphql", bytes.NewBuffer(reqPayload))
	if err != nil {
		return nil, err
	}
	return req, nil
}

func TestHttpGraphQLRequest(t *testing.T) {
	router := setupGraphQLRouter()
	body := map[string]interface{}{
		"operationName": "intrsopection",
		"query":         testutil.IntrospectionQuery,
	}
	req, err := setupRequest(http.MethodPost, "/graphql", body)
	if err != nil {
		t.Fatal(err)
	}

	_, err = router.query(req)
	if err != nil {
		t.Fatal(err)
	}
}

func TestHttpGraphQLBatchRequest(t *testing.T) {
	router := setupGraphQLRouter()
	body := []map[string]interface{}{
		map[string]interface{}{
			"operationName": "intrsopection",
			"query":         testutil.IntrospectionQuery,
		},
		map[string]interface{}{
			"operationName": "intrsopection2",
			"query":         testutil.IntrospectionQuery,
		},
	}

	req, err := setupRequest(http.MethodPost, "/graphql", body)
	if err != nil {
		t.Fatal(err)
	}

	_, err = router.query(req)
	if err != nil {
		t.Fatal(err)
	}
}

package routers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/graphql-go/graphql/testutil"
)

func setupGraphQLRouter() *GraphQLRouter {
	router := NewGraphQLRouter("test", nil, nil)
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
	router := NewGraphQLRouter("http://localhost:8080", nil, nil)
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
	router := NewGraphQLRouter("http://localhost:8080", nil, nil)
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

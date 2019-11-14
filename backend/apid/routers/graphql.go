package routers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/graphql"
)

var (
	defaultExpiration = time.Second * 30
)

type GraphQLService interface {
	Do(context.Context, graphql.QueryParams) *graphql.Result
}

// GraphQLRouter handles requests for /events
type GraphQLRouter struct {
	Service GraphQLService
}

// Mount the GraphQLRouter to a parent Router
func (r *GraphQLRouter) Mount(parent *mux.Router) {
	parent.HandleFunc("/graphql", actionHandler(r.query)).Methods(http.MethodPost)
}

func (r *GraphQLRouter) query(req *http.Request) (interface{}, error) {
	// Setup context
	ctx := context.WithValue(req.Context(), corev2.NamespaceKey, "")

	// Parse request body
	var reqBody interface{}
	if err := json.NewDecoder(req.Body).Decode(&reqBody); err != nil {
		return nil, err
	}

	// If list parse each operation
	var receivedList bool
	var ops []map[string]interface{}
	switch reqBody := reqBody.(type) {
	case []interface{}:
		receivedList = true
		for _, r := range reqBody {
			if r, ok := r.(map[string]interface{}); ok {
				ops = append(ops, r)
			}
		}
	case map[string]interface{}:
		ops = append(ops, reqBody)
	default:
		return nil, errors.New("received unexpected request body")
	}

	// Execute each operation; maybe this could be done in parallel in the future.
	results := make([]interface{}, 0, len(ops))
	for _, op := range ops {
		// Extract query and variables
		query, _ := op["query"].(string)
		queryVars, _ := op["variables"].(map[string]interface{})
		skipValidate, _ := op["skip_validation"].(bool)

		// Execute given query
		result := r.Service.Do(ctx, graphql.QueryParams{
			Query:          query,
			Variables:      queryVars,
			SkipValidation: skipValidate,
		})
		results = append(results, result)
		if len(result.Errors) > 0 {
			logger.
				WithField("errors", result.Errors).
				Error("error(s) occurred while executing GraphQL operation")
		}
	}

	if receivedList {
		return results, nil
	}
	return results[0], nil
}

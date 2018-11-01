package routers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	graphql "github.com/sensu/sensu-go/backend/apid/graphql"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	graphqlservice "github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
)

// GraphQLRouter handles requests for /events
type GraphQLRouter struct {
	service *graphqlservice.Service
}

// NewGraphQLRouter instantiates new events controller
func NewGraphQLRouter(store store.Store, bus messaging.MessageBus, getter types.QueueGetter) *GraphQLRouter {
	service, err := graphql.NewService(graphql.ServiceConfig{
		Store:       store,
		Bus:         bus,
		QueueGetter: getter,
	})
	if err != nil {
		logger.WithError(err).Panic("unable to configure graphql service")
	}
	return &GraphQLRouter{service}
}

// Mount the GraphQLRouter to a parent Router
func (r *GraphQLRouter) Mount(parent *mux.Router) {
	parent.HandleFunc("/graphql", actionHandler(r.query)).Methods(http.MethodPost)
}

func (r *GraphQLRouter) query(req *http.Request) (interface{}, error) {
	ctx := req.Context()

	// reset org & env keys to empty state so that all resources are queryable.
	ctx = context.WithValue(ctx, types.NamespaceKey, "")

	// Parse request body
	var reqBody interface{}
	if err := json.NewDecoder(req.Body).Decode(&reqBody); err != nil {
		return nil, err
	}

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

	results := make([]interface{}, 0, len(ops))
	for _, op := range ops {
		// Extract query and variables
		query, _ := op["query"].(string)
		queryVars, _ := op["variables"].(map[string]interface{})

		// Execute given query
		result := r.service.Do(ctx, query, queryVars)
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

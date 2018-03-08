package routers

import (
	"context"
	"encoding/json"
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
	ctx = context.WithValue(ctx, types.OrganizationKey, "")
	ctx = context.WithValue(ctx, types.EnvironmentKey, "")

	// Parse request body
	rBody := map[string]interface{}{}
	if err := json.NewDecoder(req.Body).Decode(&rBody); err != nil {
		return nil, err
	}

	// Extract query and variables
	query, _ := rBody["query"].(string)
	queryVars, _ := rBody["variables"].(map[string]interface{})

	// Execute given query
	result := r.service.Do(ctx, query, queryVars)
	if len(result.Errors) > 0 {
		logger.
			WithField("errors", result.Errors).
			Errorf("error(s) occurred while executing GraphQL operation")
	}

	return result, nil
}

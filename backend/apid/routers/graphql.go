package routers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/api"
	"github.com/sensu/sensu-go/backend/apid/actions"
	graphql "github.com/sensu/sensu-go/backend/apid/graphql"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

var (
	defaultExpiration = time.Second * 30
)

// GraphQLRouter handles requests for /events
type GraphQLRouter struct {
	service *graphql.Service
	store   store.Store
}

// NewGraphQLRouter instantiates new events controller
func NewGraphQLRouter(store store.Store, eventStore store.EventStore, auth authorization.Authorizer, qGetter types.QueueGetter, bus messaging.MessageBus) *GraphQLRouter {
	service, err := graphql.NewService(graphql.ServiceConfig{
		AssetClient:       api.NewAssetClient(store, auth),
		CheckClient:       api.NewCheckClient(store, actions.NewCheckController(store, qGetter), auth),
		EntityClient:      api.NewEntityClient(store, eventStore, auth),
		EventClient:       api.NewEventClient(eventStore, auth, bus),
		EventFilterClient: api.NewEventFilterClient(store, auth),
		HandlerClient:     api.NewHandlerClient(store, auth),
		MutatorClient:     api.NewMutatorClient(store, auth),
		SilencedClient:    api.NewSilencedClient(store, auth),
		NamespaceClient:   api.NewNamespaceClient(store, auth),
		HookClient:        api.NewHookConfigClient(store, auth),
		UserClient:        api.NewUserClient(store, auth),
		RBACClient:        api.NewRBACClient(store, auth),
		GenericClient:     &api.GenericClient{Store: store, Auth: auth},
	})
	if err != nil {
		logger.WithError(err).Panic("unable to configure graphql service")
	}

	router := GraphQLRouter{
		service: service,
		store:   store,
	}
	return &router
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

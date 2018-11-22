package routers

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	graphql "github.com/sensu/sensu-go/backend/apid/graphql"
	"github.com/sensu/sensu-go/backend/apid/graphql/restclient"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
	graphqlservice "github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
)

var (
	defaultExpiration = time.Millisecond * 2500
)

// GraphQLRouter handles requests for /events
type GraphQLRouter struct {
	service *graphqlservice.Service
}

// NewGraphQLRouter instantiates new events controller
func NewGraphQLRouter(apiURL string, tls *tls.Config) *GraphQLRouter {
	factory := restclient.NewClientFactory(apiURL, tls)
	service, err := graphql.NewService(graphql.ServiceConfig{
		ClientFactory: factory,
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
	// Setup context
	ctx := req.Context()
	ctx = context.WithValue(ctx, types.NamespaceKey, "")

	// Create a short-lived access token for the duration of the request and lift
	// it into the context.
	accessToken := jwt.ExtractBearerToken(req)
	if newToken, err := createShortLivedToken(accessToken); err == nil {
		ctx = context.WithValue(ctx, types.AccessTokenString, newToken)
	}

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

func createShortLivedToken(token string) (string, error) {
	accessToken, err := jwt.ValidateToken(token)
	if accessToken == nil || err != nil {
		return "", err
	}

	claims, err := jwt.GetClaims(accessToken)
	if err != nil {
		return "", err
	}

	claims.StandardClaims.ExpiresAt = time.Now().Add(defaultExpiration).Unix()
	_, token, err = jwt.NewAccessTokenWithClaims(claims)
	if err != nil {
		return "", err
	}

	return token, err
}

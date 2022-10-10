package routers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/graphql"
)

var (
	graphqlDefaultTimeout = 10 * time.Second
)

type GraphQLService interface {
	Do(context.Context, graphql.QueryParams) *graphql.Result
}

// GraphQLRouter handles requests for /events
type GraphQLRouter struct {
	Service GraphQLService
	Timeout time.Duration
}

// Mount the GraphQLRouter to a parent Router
func (r *GraphQLRouter) Mount(parent *mux.Router) {
	parent.HandleFunc("/graphql", actionHandler(r.query)).Methods(http.MethodPost)
}

func (r *GraphQLRouter) query(req *http.Request) (interface{}, error) {
	timeout := r.Timeout
	if timeout == 0 {
		timeout = graphqlDefaultTimeout
	}

	// Setup context
	ctx := context.WithValue(req.Context(), corev2.NamespaceKey, "")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

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

	claims := jwt.GetClaimsFromContext(req.Context())

	// Execute each operation; maybe this could be done in parallel in the future.
	results := make([]interface{}, 0, len(ops))
	var timedOut bool
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
		results = append(results, map[string]interface{}{
			"data":   result.Data,
			"errors": result.Errors,
			"auth":   claims != nil,
		})
		if result.HasErrors() {
			logger.
				WithField("errors", result.Errors).
				Error("error(s) occurred while executing GraphQL operation")

			// When the default graphql-go execuctor encounters an exceeded
			// deadline it will wrap the error. Since client's typically
			// consider service errors as equivelant to an unrecoverable server
			// error we instead capture the error and respond with a 504 status.
			if err := result.Errors[0]; err.Message == context.DeadlineExceeded.Error() {
				timedOut = true
				break
			}
		}
	}

	if timedOut {
		return results, actions.NewErrorf(actions.DeadlineExceeded)
	}
	if receivedList {
		return results, nil
	}
	return results[0], nil
}

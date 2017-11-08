package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	graphql "github.com/sensu/sensu-go/backend/apid/graphql"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"golang.org/x/net/context"
)

// GraphController defines the fields required by GraphController.
type GraphController struct {
	Store store.Store
}

// Register should define an association between HTTP routes and their
// respective handlers defined within this Controller.
func (c *GraphController) Register(r *mux.Router) {
	r.HandleFunc("/graphql", c.query).Methods(http.MethodPost)
}

// query handles GraphQL queries
func (c *GraphController) query(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Lift Etcd store into context so that resolvers may query and reset org &
	// env keys to empty state so that all resources are queryable.
	ctx = context.WithValue(ctx, types.OrganizationKey, "")
	ctx = context.WithValue(ctx, types.EnvironmentKey, "")
	ctx = context.WithValue(ctx, types.StoreKey, c.Store)

	// Read request body
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(
			writer,
			"request body must be valid JSON",
			http.StatusBadRequest,
		)
		return
	}
	defer r.Body.Close()

	// Parse request body
	rBody := map[string]interface{}{}
	err = json.Unmarshal(bodyBytes, &rBody)
	if err != nil {
		http.Error(
			writer,
			"request body must be valid JSON",
			http.StatusBadRequest,
		)
		return
	}

	// Extract query and variables
	query, _ := rBody["query"].(string)
	queryVars, _ := rBody["variables"].(map[string]interface{})

	// Execute given query
	result := graphql.Execute(ctx, query, &queryVars)
	if len(result.Errors) > 0 {
		logger.
			WithField("errors", result.Errors).
			Errorf("error(s) occurred while executing GraphQL operation")
	}

	// Marshal result of query execution
	rJSON, err := json.Marshal(result)
	if err != nil {
		http.Error(
			writer,
			"unknown error occured while marshalling response",
			http.StatusInternalServerError,
		)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(writer, "%s", rJSON)
}

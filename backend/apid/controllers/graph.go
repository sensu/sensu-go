package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	graphql "github.com/sensu/sensu-go/graphql"
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

// many handles requests to /info
func (c *GraphController) query(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx = context.WithValue(ctx, types.OrganizationKey, "")
	ctx = context.WithValue(ctx, types.EnvironmentKey, "")
	ctx = context.WithValue(ctx, types.StoreKey, c.Store)

	// Fake being authenticated for demoing
	actor := authorization.Actor{
		Name: "admin",
		Rules: []types.Rule{{
			Type:         "*",
			Environment:  "*",
			Organization: "*",
			Permissions:  types.RuleAllPerms,
		}},
	}
	ctx = context.WithValue(ctx, types.AuthorizationActorKey, actor)

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	rBody := map[string]interface{}{}
	err = json.Unmarshal(bodyBytes, &rBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query, _ := rBody["query"].(string) // TODO: Safer way to do this?
	queryVars, _ := rBody["variables"].(map[string]interface{})
	result := graphql.Execute(ctx, query, &queryVars)
	if len(result.Errors) > 0 {
		logger.
			WithField("errors", result.Errors).
			Errorf("error(s) occurred while executing GraphQL operation")
	}

	rJSON, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s", rJSON)
}

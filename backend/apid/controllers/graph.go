package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/graphql-go/graphql"
	graphqlast "github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/relay"
	"github.com/sensu/sensu-go/backend/authorization"
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

	res := execQuery(ctx, rBody["query"].(string))
	if len(res.Errors) > 0 {
		logger.
			WithField("errors", res.Errors).
			Errorf("failed to execute graphql operation")

		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	rJSON, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s", rJSON)
}

func execQuery(ctx context.Context, queryStr string) *graphql.Result {
	params := graphql.Params{
		Schema:        GraphqlSchema,
		RequestString: queryStr,
		Context:       ctx,
	}

	if logger.Logger.Level >= logrus.DebugLevel {
		re := regexp.MustCompile(`\s+`) // TODO move to init
		formattedQuery := queryStr
		formattedQuery = strings.Replace(formattedQuery, "\n", " ", -1)
		formattedQuery = re.ReplaceAllLiteralString(formattedQuery, " ")
		logger.WithField("query", formattedQuery).Debug("executing GraphQL query")
	}

	return graphql.Do(params)
}

// GraphqlSchema ...
var GraphqlSchema graphql.Schema

func init() {
	var err error
	var checkEventType *graphql.Object
	var checkConfigType *graphql.Object
	var metricEventType *graphql.Object
	var entityType *graphql.Object
	var userType *graphql.Object

	//
	// Relay

	nodeDefinitions := relay.NewNodeDefinitions(relay.NodeDefinitionsConfig{
		IDFetcher: func(id string, info graphql.ResolveInfo, ctx context.Context) (interface{}, error) {
			// resolve id from global id
			gidComponents := relay.FromGlobalID(id)
			store := ctx.Value(types.StoreKey).(store.Store)

			// based on id and its type, return the object
			switch gidComponents.Type {
			case "CheckEvent":
				// TODO
				return types.FixtureEvent("a", "b"), nil
			case "MetricEvent":
				// TODO
				return types.FixtureEvent("a", "b"), nil
			case "Entity":
				entity, err := store.GetEntityByID(ctx, gidComponents.ID)
				return entity, err
			case "Check":
				check, err := store.GetCheckConfigByName(ctx, gidComponents.ID)
				return check, err
			case "User":
				user, err := store.GetUser(gidComponents.ID)
				return user, err
			default:
				return nil, errors.New("Unknown node type")
			}
		},
		TypeResolve: func(p graphql.ResolveTypeParams) *graphql.Object {
			// based on the type of the value, return GraphQLObjectType
			switch p.Value.(type) {
			case *types.Event:
				return checkEventType
			case *types.Entity:
				return entityType
			default:
				return nil
			}
		},
	})

	//
	// Interface

	multitenantResource := graphql.NewInterface(graphql.InterfaceConfig{
		Name:        "MultitenantResource",
		Description: "A resource that belong to an organization and environment",
		Fields: graphql.Fields{
			"environment": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "The environment the resource belongs to.",
			},
			"organization": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "The organization the resource belongs to.",
			},
		},
		ResolveType: func(p graphql.ResolveTypeParams) *graphql.Object {
			if _, ok := p.Value.(types.Entity); ok {
				return entityType
			}
			return nil
		},
	})

	//
	// Test Scalar

	timeScalar := graphql.NewScalar(graphql.ScalarConfig{
		Name:        "Time",
		Description: "The `Time` scalar type represents an instant in time",
		Serialize:   coerceTime,
		ParseValue:  coerceTime,
		ParseLiteral: func(valueAST graphqlast.Value) interface{} {
			switch valueAST := valueAST.(type) {
			case *graphqlast.IntValue:
				if intValue, err := strconv.Atoi(valueAST.Value); err == nil {
					return time.Unix(int64(intValue), 0)
				}
			case *graphqlast.StringValue:
				// TODO: Would be nice to cover
			}
			return nil
		},
	})

	userType = graphql.NewObject(graphql.ObjectConfig{
		Name: "User",
		Interfaces: []*graphql.Interface{
			nodeDefinitions.NodeInterface,
		},
		Fields: graphql.Fields{
			"id":       relay.GlobalIDField("User", nil),
			"username": &graphql.Field{Type: graphql.String},
			"disabled": &graphql.Field{Type: graphql.Boolean},
			"hasPassword": &graphql.Field{
				Type: graphql.Boolean,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					user := p.Source.(*types.User)
					return len(user.Password) > 0, nil
				},
			},
			// NOTE: Something where we'd probably want to restrict access
			"roles": &graphql.Field{Type: graphql.NewList(graphql.String)},
		},
	})

	entityType = graphql.NewObject(graphql.ObjectConfig{
		Name: "Entity",
		Interfaces: []*graphql.Interface{
			nodeDefinitions.NodeInterface,
			multitenantResource,
		},
		Fields: graphql.Fields{
			"id":               relay.GlobalIDField("Entity", nil),
			"entityID":         AliasField(graphql.String, "ID"),
			"class":            &graphql.Field{Type: graphql.String},
			"subscriptions":    &graphql.Field{Type: graphql.NewList(graphql.String)},
			"lastSeen":         &graphql.Field{Type: graphql.String},
			"deregister":       &graphql.Field{Type: graphql.Boolean},
			"keepaliveTimeout": &graphql.Field{Type: graphql.Int},
			"environment":      &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"organization":     &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		},
		IsTypeOf: func(p graphql.IsTypeOfParams) bool {
			_, ok := p.Value.(*types.Entity)
			return ok
		},
	})

	metricEventType = graphql.NewObject(graphql.ObjectConfig{
		Name: "MetricEvent",
		Fields: graphql.Fields{
			"id":        relay.GlobalIDField("MetricEvent", nil),
			"entity":    &graphql.Field{Type: entityType},
			"timestamp": &graphql.Field{Type: timeScalar},
		},
		IsTypeOf: func(p graphql.IsTypeOfParams) bool {
			if e, ok := p.Value.(*types.Event); ok {
				return e.Metrics != nil
			}
			return false
		},
	})

	checkConfigType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "Check",
		Description: "The `Check` object type represents  the specification of a check",
		Interfaces: []*graphql.Interface{
			nodeDefinitions.NodeInterface,
		},
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Name:        "id",
				Description: "The ID of an object",
				Type:        graphql.NewNonNull(graphql.ID),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					check := p.Source.(*types.CheckConfig)
					return relay.ToGlobalID("Check", check.Name), nil
				},
			},
			"name":          &graphql.Field{Type: graphql.String},
			"interval":      &graphql.Field{Type: graphql.Int},
			"subscriptions": &graphql.Field{Type: graphql.NewList(graphql.String)},
			"command":       &graphql.Field{Type: graphql.String},
			"handlers":      &graphql.Field{Type: graphql.NewList(graphql.String)},
			"runtimeAssets": &graphql.Field{Type: graphql.NewList(graphql.String)},
			"environment":   &graphql.Field{Type: graphql.String},
			"organization":  &graphql.Field{Type: graphql.String},
		},
	})

	checkEventType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "CheckEvent",
		Description: "A check result",
		Interfaces: []*graphql.Interface{
			nodeDefinitions.NodeInterface,
		},
		Fields: graphql.Fields{
			"id":        relay.GlobalIDField("CheckEvent", nil),
			"timestamp": &graphql.Field{Type: timeScalar},
			"entity":    &graphql.Field{Type: entityType},
			"output":    AliasField(graphql.String, "Check", "Output"),
			"status":    AliasField(graphql.Int, "Check", "Status"),
			"issued":    AliasField(timeScalar, "Check", "Issued"),
			"executed":  AliasField(timeScalar, "Check", "Executed"),
			"config":    AliasField(checkConfigType, "Check", "Config"),
		},
		IsTypeOf: func(p graphql.IsTypeOfParams) bool {
			if e, ok := p.Value.(*types.Event); ok {
				return e.Check != nil
			}
			return false
		},
	})

	//
	// Test Union

	eventType := graphql.NewUnion(graphql.UnionConfig{
		Name:        "Event",
		Description: "???", // TODO: ???
		Types: []*graphql.Object{
			checkEventType,
			metricEventType,
		},
		ResolveType: func(p graphql.ResolveTypeParams) *graphql.Object {
			// TODO
			if event, ok := p.Value.(*types.Event); ok {
				if event.Check != nil {
					return checkEventType
				} else if event.Metrics != nil {
					return metricEventType
				}
			}

			return nil
		},
	})

	//
	// Connections

	entityConnectionDef := relay.ConnectionDefinitions(relay.ConnectionConfig{
		Name:     "Entity",
		NodeType: entityType,
	})

	checkConnectionDef := relay.ConnectionDefinitions(relay.ConnectionConfig{
		Name:     "Check",
		NodeType: checkConfigType,
	})

	viewerType := graphql.NewObject(graphql.ObjectConfig{
		Name:        "Viewer",
		Description: "describes resources available to the curr1nt user",
		Fields: graphql.Fields{
			"entities": &graphql.Field{
				Type: entityConnectionDef.ConnectionType,
				Args: relay.ConnectionArgs,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					args := relay.NewConnectionArguments(p.Args)

					store := p.Context.Value(types.StoreKey).(store.Store)
					entities, err := store.GetEntities(p.Context)

					resources := []interface{}{}
					for _, entity := range entities {
						resources = append(resources, entity)
					}

					return relay.ConnectionFromArray(resources, args), err
				},
			},
			"user": &graphql.Field{
				Type: userType,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					ctx := p.Context
					actor := ctx.Value(types.AuthorizationActorKey).(authorization.Actor)
					store := ctx.Value(types.StoreKey).(store.Store)

					user, err := store.GetUser(actor.Name)
					return user, err
				},
			},
			"events": &graphql.Field{
				Type: graphql.NewList(eventType),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					store := p.Context.Value(types.StoreKey).(store.Store)
					events, err := store.GetEvents(p.Context)
					return events, err
				},
			},
			"checks": &graphql.Field{
				Type: checkConnectionDef.ConnectionType,
				Args: relay.ConnectionArgs,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					args := relay.NewConnectionArguments(p.Args)

					store := p.Context.Value(types.StoreKey).(store.Store)
					checks, err := store.GetCheckConfigs(p.Context)

					resources := []interface{}{}
					for _, check := range checks {
						resources = append(resources, check)
					}

					return relay.ConnectionFromArray(resources, args), err
				},
			},
		},
	})

	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"node": nodeDefinitions.NodeField,
			"viewer": &graphql.Field{
				Type: viewerType,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return 1, nil // TODO? User? Viewer warpper type?
				},
			},
		},
	})
	GraphqlSchema, err = graphql.NewSchema(graphql.SchemaConfig{
		Query: queryType,
		// Mutation: mutationType,
	})
	if err != nil {
		logger.WithError(err).Fatal("failed to create new schema")
	}
}

func coerceTime(value interface{}) interface{} {
	switch value := value.(type) {
	case time.Time:
		return value.Format(time.RFC1123Z)
	case int64: // TODO: Too naive
		return coerceTime(time.Unix(value, 0))
	}

	return nil
}

// AliasField TODO: ...
func AliasField(T graphql.Output, fNames ...string) *graphql.Field {
	return &graphql.Field{
		Type: T,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			fVal := reflect.ValueOf(p.Source)
			for _, fName := range fNames {
				fVal = reflect.Indirect(fVal)
				fVal = fVal.FieldByName(fName)
			}
			return fVal.Interface(), nil
		},
	}
}

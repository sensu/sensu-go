package graphqlschema

import (
	"github.com/graphql-go/graphql"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/graphql/globalid"
	"github.com/sensu/sensu-go/graphql/relay"
	"github.com/sensu/sensu-go/types"
)

var checkConfigType *graphql.Object
var checkConfigConnection *relay.ConnectionDefinitions

func init() {
	initNodeInterface()
	initCheckConfigType()
	initCheckConfigConnection()

	nodeResolver := newCheckConfigNodeResolver()
	nodeRegister.RegisterResolver(nodeResolver)
}

func initCheckConfigType() {
	if checkConfigType != nil {
		return
	}

	checkConfigType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "CheckConfig",
		Description: "Represents the specification of a check",
		Interfaces: graphql.InterfacesThunk(func() []*graphql.Interface {
			return []*graphql.Interface{
				nodeInterface,
				multitenantInterface,
			}
		}),
		Fields: graphql.FieldsThunk(func() graphql.Fields {
			return graphql.Fields{
				"id": &graphql.Field{
					Name:        "id",
					Description: "The ID of an object",
					Type:        graphql.NewNonNull(graphql.ID),
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						idComponents := globalid.CheckTranslator.Encode(p.Source)
						return idComponents.String(), nil
					},
				},
				"name": &graphql.Field{
					Description: "Name is the unique identifier for a check",
					Type:        graphql.String,
				},
				"interval": &graphql.Field{
					Description: "Interval is the interval, in seconds, at which the check should be run",
					Type:        graphql.Int,
				},
				"subscriptions": &graphql.Field{
					Description: "Subscriptions is the list of subscribers for the check",
					Type:        graphql.NewList(graphql.String),
				},
				"command": &graphql.Field{
					Description: "Command is the command to be executed",
					Type:        graphql.String,
				},
				"handlerNames": &graphql.Field{
					Description: "Handlers are the event handler for the check (incidents and/or metrics)",
					Type:        graphql.NewList(graphql.String),
					Resolve:     AliasResolver("handlers"),
				},
				// "handlers": &graphql.Field{
				// 	Description: "Handlers are the event handler for the check (incidents and/or metrics)",
				// 	Type:        graphql.NewList(handlerType),
				// 	Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				// 		abilities := authorization.Handlers.WithContext(p.Context)
				// 		check, ok := p.Source.(*types.CheckConfig)
				// 		if !ok {
				// 			return nil, errors.New("source object is not an Event")
				// 		}
				//
				// 		store := p.Context.Value(types.StoreKey).(store.HandlerStore)
				// 		handlers, err := store.GetHandlers(p.Context)
				// 		if err != nil {
				// 			return nil, err
				// 		}
				//
				// 		results := []interface{}{}
				// 		for _, handler := range handlers {
				// 			for _, hName := range check.Handlers {
				// 				// Reject any handlers that are not assoicated with the check or
				// 				// handlers where the user does not have access to read.
				// 				if handler.Name == hName && abilities.CanRead(handler) {
				// 					results = append(results, handler)
				// 				}
				// 			}
				// 		}
				//
				// 		return results, nil
				// 	},
				// },
				// TODO: Implement w/ associated types
				// "runtimeAssets": &graphql.Field{
				// 	Description: "RuntimeAssets are a list of assets required to execute check",
				// 	Type:        graphql.NewList(assetType),
				// },
				// TODO: Should use type and not string
				"environment": &graphql.Field{
					Description: "Environment indicates to which env a check belongs to",
					Type:        graphql.NewNonNull(graphql.String),
					// Type:        graphql.NewNonNull(organizationType),
				},
				"organization": &graphql.Field{
					Description: "Environment indicates to which env a check belongs to",
					Type:        graphql.NewNonNull(graphql.String),
					// Type:        graphql.NewNonNull(environmentType),
				},
			}
		}),
		IsTypeOf: func(p graphql.IsTypeOfParams) bool {
			_, ok := p.Value.(*types.CheckConfig)
			return ok
		},
	})
}

func initCheckConfigConnection() {
	if checkConfigConnection != nil {
		return
	}
	checkConfigConnection = relay.NewConnectionDefinition(relay.ConnectionConfig{
		Name:     "CheckConfig",
		NodeType: checkConfigType,
	})
}

func newCheckConfigNodeResolver() relay.NodeResolver {
	return relay.NodeResolver{
		Object:     checkConfigType,
		Translator: globalid.CheckTranslator,
		Resolve: func(p relay.NodeResolverParams) (interface{}, error) {
			components := p.IDComponents.(globalid.NamedComponents)
			store := p.Context.Value(types.StoreKey).(store.CheckConfigStore)
			record, err := store.GetCheckConfigByName(p.Context, components.Name())
			if err != nil {
				return nil, err
			}

			abilities := authorization.Checks.WithContext(p.Context)
			if abilities.CanRead(record) {
				return record, err
			}
			return nil, err
		},
	}
}

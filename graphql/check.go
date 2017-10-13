package graphqlschema

import (
	"github.com/graphql-go/graphql"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/graphql/globalid"
	"github.com/sensu/sensu-go/graphql/relay"
	"github.com/sensu/sensu-go/types"
	"golang.org/x/net/context"
)

var checkConfigType *graphql.Object
var checkConfigConnection *relay.ConnectionDefinitions

func init() {
	initNodeInterface()
	initCheckConfigType()
	initCheckConfigConnection()
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
						idComponents := globalid.CheckResource.Encode(p.Source)
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
				"handlers": &graphql.Field{
					Description: "Handlers are the event handler for the check (incidents and/or metrics)",
					Type:        graphql.NewList(graphql.String),
				},
				"runtimeAssets": &graphql.Field{
					Description: "RuntimeAssets are a list of assets required to execute check",
					Type:        graphql.NewList(graphql.String),
				},
				"environment": &graphql.Field{
					Description: "Environment indicates to which env a check belongs to",
					Type:        graphql.NewNonNull(graphql.String),
				},
				"organization": &graphql.Field{
					Description: "Environment indicates to which env a check belongs to",
					Type:        graphql.NewNonNull(graphql.String),
				},
			}
		}),
	})

	nodeRegister.RegisterResolver(relay.NodeResolver{
		Object:     checkConfigType,
		Translator: globalid.CheckResource,
		Resolve: func(ctx context.Context, c globalid.Components) (interface{}, error) {
			components := c.(globalid.NamedComponents)
			store := ctx.Value(types.StoreKey).(store.CheckConfigStore)

			// TODO: Filter out unauthorized results
			record, err := store.GetCheckConfigByName(ctx, components.Name())
			return record, err
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

package graphqlschema

import (
	"github.com/graphql-go/graphql"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/graphql/globalid"
	"github.com/sensu/sensu-go/graphql/relay"
	"github.com/sensu/sensu-go/types"
)

var entityType *graphql.Object
var entityConnection *relay.ConnectionDefinitions

func initEntityType() {
	if entityType != nil {
		return
	}

	networkInterfaceType := graphql.NewObject(graphql.ObjectConfig{
		Name: "NetworkInterface",
		Fields: graphql.FieldsThunk(func() graphql.Fields {
			return graphql.Fields{
				"name": &graphql.Field{
					Description: "The name of the interface",
					Type:        graphql.String,
				},
				"mac": &graphql.Field{
					Description: "The MAC address of the interface",
					Type:        graphql.String,
				},
				"addresses": &graphql.Field{
					Description: "The addresses that belong to the interface",
					Type:        graphql.NewList(graphql.String),
				},
			}
		}),
	})

	networkType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Network",
		Fields: graphql.FieldsThunk(func() graphql.Fields {
			return graphql.Fields{
				"interfaces": &graphql.Field{
					Description: "A list of all the network interfaces",
					Type:        graphql.NewList(networkInterfaceType),
				},
			}
		}),
	})

	systemType := graphql.NewObject(graphql.ObjectConfig{
		Name: "System",
		Fields: graphql.FieldsThunk(func() graphql.Fields {
			return graphql.Fields{
				"hostname": &graphql.Field{
					Description: "The hostname of the system",
					Type:        graphql.NewNonNull(graphql.String),
				},
				"os": &graphql.Field{
					Description: "The operating system of the system",
					Type:        graphql.String,
				},
				"platform": &graphql.Field{
					Description: "The platform of the system",
					Type:        graphql.String,
				},
				"platformFamily": &graphql.Field{
					Description: "The platform family of the system",
					Type:        graphql.String,
				},
				"platformVersion": &graphql.Field{
					Description: "The version of the platform for the system",
					Type:        graphql.String,
				},
				"network": &graphql.Field{
					Description: "The network interfaces on the system",
					Type:        networkType,
				},
			}
		}),
	})

	entityType = graphql.NewObject(graphql.ObjectConfig{
		Name: "Entity",
		Interfaces: graphql.InterfacesThunk(func() []*graphql.Interface {
			return []*graphql.Interface{
				nodeInterface,
				multitenantInterface,
			}
		}),
		Fields: graphql.FieldsThunk(func() graphql.Fields {
			return graphql.Fields{
				"id": &graphql.Field{
					Description: "The ID of an object",
					Type:        graphql.NewNonNull(graphql.ID),
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						idComponents := globalid.EntityTranslator.Encode(p.Source)
						return idComponents.String(), nil
					},
				},
				// TODO: replce with alias resolve
				// "entityID":         AliasField(graphql.String, "ID"),
				"class": &graphql.Field{
					Description: "The type of entity",
					Type:        graphql.String,
				},
				"system": &graphql.Field{
					Description: "The system information of the entity",
					Type:        systemType,
				},
				"subscriptions": &graphql.Field{
					Description: "A list of the entity subscriptions",
					Type:        graphql.NewList(graphql.String),
				},
				"lastSeen": &graphql.Field{
					Description: "The last time the backend recieved a keepalive from the entity",
					Type:        graphql.String,
				},
				"deregister": &graphql.Field{
					Description: "If a deregisteation event should be created on the agent process stopping",
					Type:        graphql.Boolean,
				},
				// TODO: figure out what this actually does
				"keepaliveTimeout": &graphql.Field{
					Description: "",
					Type:        graphql.Int,
				},
				"environment": &graphql.Field{
					Description: "The environment the entity belongs to",
					Type:        graphql.NewNonNull(graphql.String),
				},
				"organization": &graphql.Field{
					Description: "The organization the entity belongs to",
					Type:        graphql.NewNonNull(graphql.String),
				},
				// TODO: write description and resolve
				"user": &graphql.Field{
					Description: "???",
					Type:        userType,
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						entity := p.Source.(*types.Entity)
						store := p.Context.Value(types.StoreKey).(store.UserStore)
						record, err := store.GetUser(entity.User)
						if err != nil {
							return nil, err
						}

						abilities := authorization.Users.WithContext(p.Context)
						if abilities.CanRead(record) {
							return record, err
						}
						return nil, nil
					},
				},
			}
		}),
		IsTypeOf: func(p graphql.IsTypeOfParams) bool {
			_, ok := p.Value.(*types.Entity)
			return ok
		},
	})
}

func initEntityConnection() {
	if entityConnection != nil {
		return
	}

	entityConnection = relay.NewConnectionDefinition(relay.ConnectionConfig{
		Name:     "Entity",
		NodeType: entityType,
	})
}

func newEntityNodeResolver() relay.NodeResolver {
	return relay.NodeResolver{
		Object:     entityType,
		Translator: globalid.EntityTranslator,
		Resolve: func(p relay.NodeResolverParams) (interface{}, error) {
			components := p.IDComponents.(globalid.NamedComponents)
			store := p.Context.Value(types.StoreKey).(store.EntityStore)
			record, err := store.GetEntityByID(p.Context, components.Name())
			if err != nil {
				return nil, err
			}

			abilities := authorization.Entities.WithContext(p.Context)
			if abilities.CanRead(record) {
				return record, nil
			}
			return nil, nil
		},
	}
}

func init() {
	initNodeInterface()
	initEntityType()
	initEntityConnection()

	nodeResolver := newEntityNodeResolver()
	nodeRegister.RegisterResolver(nodeResolver)
}

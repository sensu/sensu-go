package graphqlschema

import (
	"github.com/graphql-go/graphql"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/apid/graphql/relay"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

var hookConfigType *graphql.Object
var hookConfigConnection *relay.ConnectionDefinitions

func init() {
	initNodeInterface()
	initHookConfigType()
	initHookConfigConnection()

	nodeResolver := newHookConfigNodeResolver()
	nodeRegister.RegisterResolver(nodeResolver)
}

func initHookConfigType() {
	if hookConfigType != nil {
		return
	}

	hookConfigType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "HookConfig",
		Description: "Represents the specification of a hook",
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
						idComponents := globalid.HookTranslator.Encode(p.Source)
						return idComponents.String(), nil
					},
				},
				"name": &graphql.Field{
					Description: "Name is the unique identifier for a hook",
					Type:        graphql.NewNonNull(graphql.String),
				},
				"timeout": &graphql.Field{
					Description: "Timeout is the timeout, in seconds, at which the hook has to run",
					Type:        graphql.NewNonNull(graphql.Int),
				},
				"command": &graphql.Field{
					Description: "Command is the command to be executed",
					Type:        graphql.String,
				},
				"stdin": &graphql.Field{
					Description: "Indicates if hook requests have stdin enabled",
					Type:        graphql.Boolean,
				},
				"environment": &graphql.Field{
					Description: "Environment indicates to which env a hook belongs to",
					Type:        graphql.NewNonNull(graphql.String),
				},
				"organization": &graphql.Field{
					Description: "Environment indicates to which env a hook belongs to",
					Type:        graphql.NewNonNull(graphql.String),
				},
			}
		}),
		IsTypeOf: func(p graphql.IsTypeOfParams) bool {
			_, ok := p.Value.(*types.HookConfig)
			return ok
		},
	})
}

func initHookConfigConnection() {
	if hookConfigConnection != nil {
		return
	}
	hookConfigConnection = relay.NewConnectionDefinition(relay.ConnectionConfig{
		Name:     "HookConfig",
		NodeType: hookConfigType,
	})
}

func newHookConfigNodeResolver() relay.NodeResolver {
	return relay.NodeResolver{
		Object:     hookConfigType,
		Translator: globalid.HookTranslator,
		Resolve: func(p relay.NodeResolverParams) (interface{}, error) {
			components := p.IDComponents.(globalid.NamedComponents)
			store := p.Context.Value(types.StoreKey).(store.HookConfigStore)
			controller := actions.NewHookController(store)

			record, err := controller.Find(p.Context, components.Name())
			if err == nil {
				return record, nil
			}

			s, ok := actions.StatusFromError(err)
			if ok && s == actions.NotFound {
				return nil, nil
			}
			return nil, err
		},
	}
}

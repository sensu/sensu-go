package graphqlschema

import (
	"github.com/graphql-go/graphql"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/graphql/globalid"
	"github.com/sensu/sensu-go/graphql/relay"
	"github.com/sensu/sensu-go/types"
)

var userType *graphql.Object

func init() {
	userType = graphql.NewObject(graphql.ObjectConfig{
		Name: "User",
		Interfaces: graphql.InterfacesThunk(func() []*graphql.Interface {
			return []*graphql.Interface{
				nodeInterface,
			}
		}),
		Fields: graphql.FieldsThunk(func() graphql.Fields {
			return graphql.Fields{
				"id": &graphql.Field{
					Description: "The ID of an object",
					Type:        graphql.NewNonNull(graphql.ID),
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						idComponents := globalid.UserTranslator.Encode(p.Source)
						return idComponents.String(), nil
					},
				},
				"username": &graphql.Field{
					Type:        graphql.String,
					Description: "The unique identifier of the user",
				},
				"disabled": &graphql.Field{
					Type:        graphql.Boolean,
					Description: "Whether or not the user's is active",
				},
				// "roles": &graphql.Field{
				// 	Type:        graphql.NewList(roleType),
				// 	Description: "Roles the user holds in the system",
				// 	Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				// 		return nil, nil
				// 	},
				// },
			}
		}),
		IsTypeOf: func(p graphql.IsTypeOfParams) bool {
			_, ok := p.Value.(*types.User)
			return ok
		},
	})

	nodeResolver := newUserNodeResolver()
	nodeRegister.RegisterResolver(nodeResolver)
}

func newUserNodeResolver() relay.NodeResolver {
	return relay.NodeResolver{
		Object:     userType,
		Translator: globalid.UserTranslator,
		Resolve: func(p relay.NodeResolverParams) (interface{}, error) {
			components := p.IDComponents.(globalid.NamedComponents)
			store := p.Context.Value(types.StoreKey).(store.UserStore)
			record, err := store.GetUser(components.Name())
			if err != nil {
				return nil, err
			}

			abilities := authorization.Users.WithContext(p.Context)
			if abilities.CanRead(record) {
				return record, err
			}
			return nil, nil
		},
	}
}

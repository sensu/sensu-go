package graphqlschema

import (
	"github.com/graphql-go/graphql"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/graphql/globalid"
	"github.com/sensu/sensu-go/graphql/relay"
	"github.com/sensu/sensu-go/types"
	"golang.org/x/net/context"
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
						idComponents := globalid.UserResource.Encode(p.Source)
						return idComponents.String(), nil
					},
				},
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
			}
		}),
		IsTypeOf: func(p graphql.IsTypeOfParams) bool {
			_, ok := p.Value.(*types.User)
			return ok
		},
	})

	nodeRegister.RegisterResolver(relay.NodeResolver{
		Object:     userType,
		Translator: globalid.UserResource,
		Resolve: func(ctx context.Context, c globalid.Components) (interface{}, error) {
			components := c.(globalid.NamedComponents)
			store := ctx.Value(types.StoreKey).(store.UserStore)

			// TODO: Filter out unauthorized results
			record, err := store.GetUser(components.Name())
			return record, err
		},
	})
}

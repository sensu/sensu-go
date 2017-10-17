package graphqlschema

import (
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/relay"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

var viewerType *graphql.Object

func init() {
	viewerType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "Viewer",
		Description: "A viewer of the system.",
		Fields: graphql.FieldsThunk(func() graphql.Fields {
			return graphql.Fields{
				"entities": &graphql.Field{
					Type:        entityConnection.ConnectionType,
					Description: "A list of entities the given viewer has read access to",
					Args:        relay.ConnectionArgs,
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						abilities := authorization.Entities.WithContext(p.Context)
						store := p.Context.Value(types.StoreKey).(store.Store)
						entities, err := store.GetEntities(p.Context)
						if err != nil {
							return nil, err
						}

						resources := []interface{}{}
						for _, entity := range entities {
							if abilities.CanRead(entity) {
								resources = append(resources, entity)
							}
						}

						args := relay.NewConnectionArguments(p.Args)
						return relay.ConnectionFromArray(resources, args), err
					},
				},
				"checks": &graphql.Field{
					Type:        checkConfigConnection.ConnectionType,
					Description: "A list of checks the given viewer has read access to",
					Args:        relay.ConnectionArgs,
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						abilities := authorization.Checks.WithContext(p.Context)
						store := p.Context.Value(types.StoreKey).(store.Store)
						checks, err := store.GetCheckConfigs(p.Context)
						if err != nil {
							return nil, err
						}

						resources := []interface{}{}
						for _, check := range checks {
							if abilities.CanRead(check) {
								resources = append(resources, check)
							}
						}

						args := relay.NewConnectionArguments(p.Args)
						return relay.ConnectionFromArray(resources, args), nil
					},
				},
				"user": &graphql.Field{
					Type:        userType,
					Description: "User account associated with the viewer.",
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						ctx := p.Context
						actor := ctx.Value(types.AuthorizationActorKey).(authorization.Actor)
						store := ctx.Value(types.StoreKey).(store.Store)

						user, err := store.GetUser(actor.Name)
						if err != nil {
							return nil, err
						}

						return user, err
					},
				},
			}
		}),
	})
}

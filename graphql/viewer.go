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
		Description: "describes resources available to the curr1nt user",
		Fields: graphql.Fields{
			"entities": &graphql.Field{
				Type: entityConnection.ConnectionType,
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
				Type: checkEventConnection.ConnectionType,
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
}

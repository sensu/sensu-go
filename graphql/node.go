package graphqlschema

import (
	"errors"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/relay"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"golang.org/x/net/context"
)

var nodeDefinitions *relay.NodeDefinitions

func init() {
	nodeDefinitions = relay.NewNodeDefinitions(relay.NodeDefinitionsConfig{
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
			// TODO
			// case *types.Event:
			// 	return checkEventType
			// case *types.Entity:
			// 	return entityType
			default:
				return nil
			}
		},
	})
}

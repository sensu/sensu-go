package graphqlschema

import (
	"github.com/graphql-go/graphql"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/graphql/globalid"
	"github.com/sensu/sensu-go/graphql/relay"
	"github.com/sensu/sensu-go/types"
	"golang.org/x/net/context"
)

var entityType *graphql.Object
var entityConnection *relay.ConnectionDefinitions

func init() {
	entityType = graphql.NewObject(graphql.ObjectConfig{
		Name: "Entity",
		Interfaces: []*graphql.Interface{
			nodeInterface,
			multitenantInterface,
		},
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Description: "The ID of an object",
				Type:        graphql.NewNonNull(graphql.ID),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					idComponents := globalid.EntityResource.Encode(p.Source)
					return idComponents.String(), nil
				},
			},
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

	nodeRegister.RegisterResolver(relay.NodeResolver{
		Object:     entityType,
		Translator: globalid.EntityResource,
		Resolve: func(ctx context.Context, c globalid.Components) (interface{}, error) {
			components := c.(globalid.NamedComponents)
			store := ctx.Value(types.StoreKey).(store.EntityStore)

			// TODO: Filter out unauthorized results
			record, err := store.GetEntityByID(ctx, components.Name())
			return record, err
		},
	})

	entityConnection = relay.NewConnectionDefinition(relay.ConnectionConfig{
		Name:     "Entity",
		NodeType: entityType,
	})
}

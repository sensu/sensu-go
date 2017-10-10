package graphqlschema

import (
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/relay"
	"github.com/sensu/sensu-go/types"
)

var entityType *graphql.Object
var entityConnection *relay.GraphQLConnectionDefinitions

func init() {
	entityType = graphql.NewObject(graphql.ObjectConfig{
		Name: "Entity",
		Interfaces: []*graphql.Interface{
			nodeDefinitions.NodeInterface,
			multitenantResource,
		},
		Fields: graphql.Fields{
			"id":               relay.GlobalIDField("Entity", nil),
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

	entityConnection = relay.ConnectionDefinitions(relay.ConnectionConfig{
		Name:     "Entity",
		NodeType: entityType,
	})
}

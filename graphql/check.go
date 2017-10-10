package graphqlschema

import (
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/relay"
	"github.com/sensu/sensu-go/types"
)

var checkConfigType *graphql.Object
var checkEventType *graphql.Object
var checkEventConnection *relay.GraphQLConnectionDefinitions

func init() {
	checkConfigType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "Check",
		Description: "The `Check` object type represents  the specification of a check",
		Interfaces: []*graphql.Interface{
			nodeDefinitions.NodeInterface,
		},
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Name:        "id",
				Description: "The ID of an object",
				Type:        graphql.NewNonNull(graphql.ID),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					check := p.Source.(*types.CheckConfig)
					return relay.ToGlobalID("Check", check.Name), nil
				},
			},
			"name":          &graphql.Field{Type: graphql.String},
			"interval":      &graphql.Field{Type: graphql.Int},
			"subscriptions": &graphql.Field{Type: graphql.NewList(graphql.String)},
			"command":       &graphql.Field{Type: graphql.String},
			"handlers":      &graphql.Field{Type: graphql.NewList(graphql.String)},
			"runtimeAssets": &graphql.Field{Type: graphql.NewList(graphql.String)},
			"environment":   &graphql.Field{Type: graphql.String},
			"organization":  &graphql.Field{Type: graphql.String},
		},
	})

	checkEventType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "CheckEvent",
		Description: "A check result",
		Interfaces: []*graphql.Interface{
			nodeDefinitions.NodeInterface,
		},
		Fields: graphql.Fields{
			"id":        relay.GlobalIDField("CheckEvent", nil),
			"timestamp": &graphql.Field{Type: timeScalar},
			"entity":    &graphql.Field{Type: entityType},
			"output":    AliasField(graphql.String, "Check", "Output"),
			"status":    AliasField(graphql.Int, "Check", "Status"),
			"issued":    AliasField(timeScalar, "Check", "Issued"),
			"executed":  AliasField(timeScalar, "Check", "Executed"),
			"config":    AliasField(checkConfigType, "Check", "Config"),
		},
		IsTypeOf: func(p graphql.IsTypeOfParams) bool {
			if e, ok := p.Value.(*types.Event); ok {
				return e.Check != nil
			}
			return false
		},
	})

	checkEventConnection = relay.ConnectionDefinitions(relay.ConnectionConfig{
		Name:     "Check",
		NodeType: checkConfigType,
	})
}

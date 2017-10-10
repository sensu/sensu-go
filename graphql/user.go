package graphqlschema

import (
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/relay"
	"github.com/sensu/sensu-go/types"
)

var userType *graphql.Object

func init() {
	userType = graphql.NewObject(graphql.ObjectConfig{
		Name: "User",
		Interfaces: []*graphql.Interface{
			nodeDefinitions.NodeInterface,
		},
		Fields: graphql.Fields{
			"id":       relay.GlobalIDField("User", nil),
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
		},
	})
}

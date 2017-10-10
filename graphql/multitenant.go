package graphqlschema

import (
	"github.com/graphql-go/graphql"
	"github.com/sensu/sensu-go/types"
)

var multitenantResource *graphql.Interface

func init() {
	multitenantResource = graphql.NewInterface(graphql.InterfaceConfig{
		Name:        "MultitenantResource",
		Description: "A resource that belong to an organization and environment",
		Fields: graphql.Fields{
			"environment": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "The environment the resource belongs to.",
			},
			"organization": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "The organization the resource belongs to.",
			},
		},
		ResolveType: func(p graphql.ResolveTypeParams) *graphql.Object {
			if _, ok := p.Value.(types.Entity); ok {
				return entityType
			}
			return nil
		},
	})
}

package graphqlschema

import (
	"github.com/graphql-go/graphql"
	"github.com/sensu/sensu-go/types"
)

var multitenantInterface *graphql.Interface

func init() {
	multitenantInterface = graphql.NewInterface(graphql.InterfaceConfig{
		Name:        "MultitenantResource",
		Description: "A resource that belong to an organization and environment",
		Fields: graphql.FieldsThunk(func() graphql.Fields {
			return graphql.Fields{
				"environment": &graphql.Field{
					// Type:        graphql.NewNonNull(environmentType),
					Type:        graphql.NewNonNull(graphql.String),
					Description: "The environment the resource belongs to.",
				},
				"organization": &graphql.Field{
					// Type:        graphql.NewNonNull(organizationType),
					Type:        graphql.NewNonNull(graphql.String),
					Description: "The organization the resource belongs to.",
				},
			}
		}),
		ResolveType: func(p graphql.ResolveTypeParams) *graphql.Object {
			switch p.Value.(type) {
			case *types.Entity:
				return entityType
			case *types.CheckConfig:
				return checkConfigType
			case *types.User:
				return userType
			}
			return nil
		},
	})
}

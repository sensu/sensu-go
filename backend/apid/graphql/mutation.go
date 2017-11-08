package graphqlschema

import (
	"github.com/graphql-go/graphql"
)

var mutationType *graphql.Object

func init() {
	mutationType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "Mutation",
		Description: "The root query for implementing GraphQL mutations.",
		Fields: graphql.FieldsThunk(func() graphql.Fields {
			return graphql.Fields{
				"createCheck":  createCheckMutation,
				"updateCheck":  updateCheckMutation,
				"destroyCheck": destroyCheckMutation,
			}
		}),
	})
}

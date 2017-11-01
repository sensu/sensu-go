package graphqlschema

import (
	"github.com/graphql-go/graphql"
)

var schemaMemo *graphql.Schema

// Schema ...
func Schema() graphql.Schema {
	if schemaMemo != nil {
		return *schemaMemo
	}

	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: queryType,
		// Mutation: mutationType,
		// Subscription: subscriptionType,
	})

	// NOTE:
	// eventType is not exposed anywhere at the moment so we explicitly append it.
	schema.AppendType(eventType)

	if err != nil {
		logEntry := logger.WithError(err)
		logEntry.Fatal("unable to configure GraphQL schema")
	}

	schemaMemo = &schema
	return schema
}

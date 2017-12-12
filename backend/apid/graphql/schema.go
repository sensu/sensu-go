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
		Query:    queryType,
		Mutation: mutationType,
		// Subscription: subscriptionType,
	})

	if err != nil {
		logEntry := logger.WithError(err)
		logEntry.Fatal("unable to configure GraphQL schema")
	}

	// NOTE:
	// eventType is not exposed anywhere at the moment so we explicitly append it.
	if err := schema.AppendType(eventType); err != nil {
		logEntry := logger.WithError(err)
		logEntry.Fatal("unable to configure GraphQL schema")
	}

	schemaMemo = &schema
	return schema
}

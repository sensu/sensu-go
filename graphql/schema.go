package graphqlschema

import "github.com/graphql-go/graphql"

var Schema graphql.Schema

func init() {
	var err error

	Schema, err = graphql.NewSchema(graphql.SchemaConfig{
		Query: queryType,
		// Mutation: mutationType,
		// Subscription: subscriptionType,
	})

	if err != nil {
		logEntry := logger.WithError(err)
		logEntry.Fatal("unable to configure GraphQL schema")
	}
}

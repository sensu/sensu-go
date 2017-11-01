package main

import (
	"encoding/json"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/testutil"
	"github.com/sensu/sensu-go/backend/apid/graphql"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	// Save JSON of full schema introspection for Babel Relay Plugin to use
	result := graphql.Do(graphql.Params{
		Schema:        graphqlschema.Schema(),
		RequestString: testutil.IntrospectionQuery,
	})
	if result.HasErrors() {
		log.Fatalf("ERROR introspecting schema: %v", result.Errors)
		return
	}

	b, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Fatalf("ERROR: %v", err)
	}

	err = ioutil.WriteFile("./data/schema.json", b, os.ModePerm)
	if err != nil {
		log.Fatalf("ERROR: %v", err)
	}
}

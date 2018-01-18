package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/graphql-go/graphql/testutil"
	"github.com/sensu/sensu-go/backend/apid/graphql"
)

func main() {
	// Save JSON of full schema introspection for Babel Relay Plugin to use
	service, err := graphql.NewService(graphql.ServiceConfig{})
	if err != nil {
		log.Fatalf("ERROR: %v", err)
		return
	}

	result := service.Do(
		context.TODO(),
		testutil.IntrospectionQuery,
		map[string]interface{}{},
	)
	if result.HasErrors() {
		log.Fatalf("ERROR introspecting schema: %v", result.Errors)
		return
	}

	b, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Fatalf("ERROR: %v", err)
		return
	}

	err = ioutil.WriteFile("./data/schema.json", b, os.ModePerm)
	if err != nil {
		log.Fatalf("ERROR: %v", err)
	}
}

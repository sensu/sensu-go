package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	testutil "github.com/graphql-go/graphql/testutil"
	core "github.com/sensu/sensu-go/backend/apid/graphql"
	"github.com/sensu/sensu-go/graphql"
)

func main() {
	core, err := core.NewService(core.ServiceConfig{})
	if err != nil {
		log.Fatalf("unable to instantiate GraphQL service: %s", err)
	}

	result := core.Target.Do(context.Background(), graphql.QueryParams{Query: testutil.IntrospectionQuery})
	if len(result.Errors) > 0 {
		log.Fatalf("error while printing schema: %s", result.Errors[0])
	}

	enc := json.NewEncoder(os.Stdout)
	if err := enc.Encode(result.Data); err != nil {
		log.Fatal(err)
	}
}

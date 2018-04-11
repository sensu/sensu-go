package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/graphql-go/graphql/testutil"
	"github.com/sensu/sensu-go/backend/apid/graphql"
	"github.com/sensu/sensu-go/testing/mockqueue"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/mock"
)

func main() {
	// use a mock for the store since we need a non-nil store to avoid issues when
	// trying to create a queue in the service
	store := &mockstore.MockStore{}
	getter := &mockqueue.Getter{}
	queue := &mockqueue.MockQueue{}
	getter.On("GetQueue", mock.Anything).Return(queue)
	// Save JSON of full schema introspection for Babel Relay Plugin to use
	service, err := graphql.NewService(graphql.ServiceConfig{Store: store, QueueGetter: getter})
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

	err = ioutil.WriteFile("./src/schema.json", b, os.ModePerm)
	if err != nil {
		log.Fatalf("ERROR: %v", err)
	}
}

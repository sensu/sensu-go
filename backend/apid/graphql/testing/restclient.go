package testing

import (
	"context"

	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/cli/client/testing"
)

type testClientFactory struct {
	client *testing.MockClient
}

func (r testClientFactory) NewWithContext(ctx context.Context) client.APIClient {
	return r.client
}

func NewClientFactory() (*testing.MockClient, testClientFactory) {
	client := testing.MockClient{}
	factory := testClientFactory{&client}

	return &client, factory
}

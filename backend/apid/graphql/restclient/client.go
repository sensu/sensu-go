package restclient

import (
	"context"

	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/cli/client/config/inmemory"
	"github.com/sensu/sensu-go/types"
)

// ClientFactory instantiates new copies of the REST API client
type ClientFactory struct {
	url string
}

// NewClientFactory instantiates new ClientFactory
func NewClientFactory(url string) *ClientFactory {
	factory := ClientFactory{
		url: url,
	}

	return &factory
}

// NewWithContext takes a context and returns new REST API client
func (c *ClientFactory) NewWithContext(ctx context.Context) client.APIClient {
	var accessToken string
	if token, ok := ctx.Value(types.AccessTokenString).(string); ok {
		accessToken = token
	}

	tokens := types.Tokens{
		Access:    accessToken,
		ExpiresAt: 10000000000000, // TODO
	}

	config := inmemory.New(c.url)
	config.SaveTokens(&tokens)
	config.SaveNamespace(types.ContextNamespace(ctx))

	return client.New(config)
}

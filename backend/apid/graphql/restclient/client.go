package main

import (
	"context"

	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/cli/client/config/inmemory"
	"github.com/sensu/sensu-go/types"
)

type ClientFactory struct {
	url string
}

func NewClientFactory(url string) *ClientFactory {
	factory := ClientFactory{
		url: url,
	}

	return &factory
}

func (c *ClientFactory) NewWithContext(ctx context.Context) client.APIClient {
	var accessToken string
	if token, ok := ctx.Value(types.AccessTokenString).(string); ok {
		accessToken = token
	}

	config := inmemory.New(c.url)
	config.SaveTokens(&types.Tokens{Access: accessToken})
	config.SaveNamespace(types.ContextNamespace(ctx))

	return client.New(config)
}

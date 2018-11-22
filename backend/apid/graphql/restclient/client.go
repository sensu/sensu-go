package restclient

import (
	"context"
	"crypto/tls"

	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/cli/client/config/inmemory"
	"github.com/sensu/sensu-go/types"
)

// CRUFT: Avoid having the client error by setting expire in the far future
const defaultExpiry = 2524636800000 // Jan 1, 2050

// ClientFactory instantiates new copies of the REST API client
type ClientFactory struct {
	url       string
	tlsConfig *tls.Config
}

// NewClientFactory instantiates new ClientFactory
func NewClientFactory(url string, tlsConfig *tls.Config) *ClientFactory {
	factory := ClientFactory{
		url:       url,
		tlsConfig: tlsConfig,
	}

	return &factory
}

// NewWithContext takes a context and returns new REST API client
func (c *ClientFactory) NewWithContext(ctx context.Context) client.APIClient {
	config := inmemory.New(c.url)
	tokens := types.Tokens{ExpiresAt: defaultExpiry}
	if token, ok := ctx.Value(types.AccessTokenString).(string); ok {
		tokens.Access = token
	}

	// The inmemory client should /never/ return an err
	if err := config.SaveTokens(&tokens); err != nil {
		panic(err)
	}

	nsp := types.ContextNamespace(ctx)
	if err := config.SaveNamespace(nsp); err != nil {
		panic(err)
	}

	client := client.New(config)
	client.SetTLSClientConfig(c.tlsConfig)

	return client
}

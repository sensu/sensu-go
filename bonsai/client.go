package bonsai

import (
	"crypto/tls"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

var logger = logrus.WithFields(logrus.Fields{
	"component": "bonsai-client",
})

// DefaultEndpointURL is the default url for bonsai assets.
const DefaultEndpointURL = "https://bonsai.sensu.io/api/v1/assets"

// Config is the configuration for bonsai.
type Config struct {
	// EndpointURL is the URL of Bonsai.
	EndpointURL string

	// TLSConfig allows overriding client TLS configuration. Should only be
	// needed for testing.
	TLSConfig *tls.Config
}

// Client specifies the client interface of a bonsai client.
type Client interface {
	FetchAsset(string, string) (*Asset, error)
	FetchAssetVersion(string, string, string) (string, error)
}

// RestClient is a REST client for Bonsai.
type RestClient struct {
	httpClient http.Client
	config     Config
}

// New builds a new client with defaults
func New(config Config) *RestClient {
	if config.EndpointURL == "" {
		config.EndpointURL = DefaultEndpointURL
	}

	client := &RestClient{config: config}

	// set http client timeout
	client.httpClient.Timeout = 15 * time.Second

	if config.TLSConfig != nil {
		transport := new(http.Transport)
		transport.TLSClientConfig = config.TLSConfig
		client.httpClient.Transport = transport
	}

	return client
}

func (c *RestClient) newGetRequest(slugs ...string) (*http.Request, error) {
	server := c.config.EndpointURL
	if !strings.HasSuffix(server, "/") {
		server = server + "/"
	}
	path := server + path.Join(slugs...)
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

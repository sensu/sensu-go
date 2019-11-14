package bonsai

import (
	"crypto/tls"
	"time"

	"github.com/go-resty/resty"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

// DefaultEndpointURL is the default url for bonsai assets.
const DefaultEndpointURL = "https://bonsai.sensu.io/api/v1/assets"

// Config is the configuration for bonsai.
type Config struct {
	EndpointURL string
}

type Client interface {
	FetchAsset(string, string) (*Asset, error)
	FetchAssetVersion(string, string, string) (string, error)
}

// RestClient wraps resty.Client
type RestClient struct {
	resty  *resty.Client
	config Config

	configured bool
}

func init() {
	logger = logrus.WithFields(logrus.Fields{
		"component": "bonsai-client",
	})
}

// New builds a new client with defaults
func New(config Config) *RestClient {
	restyInst := resty.New()

	if config.EndpointURL == "" {
		config.EndpointURL = DefaultEndpointURL
	}

	client := &RestClient{resty: restyInst, config: config}

	// set http client timeout
	restyInst.SetTimeout(15 * time.Second)

	// Standardize redirect policy
	restyInst.SetRedirectPolicy(resty.FlexibleRedirectPolicy(10))

	// JSON
	restyInst.SetHeader("Accept", "application/json")
	restyInst.SetHeader("Content-Type", "application/json")

	// logging
	w := logger.Writer()
	defer func() {
		_ = w.Close()
	}()
	restyInst.SetLogger(w)

	return client
}

// R returns new resty.Request from configured client
func (client *RestClient) R() *resty.Request {
	client.configure()
	request := client.resty.R()

	return request
}

// SetTLSClientConfig assigns client TLS config
func (client *RestClient) SetTLSClientConfig(c *tls.Config) {
	client.resty.SetTLSClientConfig(c)
}

// Reset client so that it reconfigure on next request
func (client *RestClient) Reset() {
	client.configured = false
}

func (client *RestClient) configure() {
	if client.configured {
		return
	}

	restyInst := client.resty
	config := client.config

	// Set URL
	restyInst.SetHostURL(config.EndpointURL)

	client.configured = true
}

//go:build fips140
// +build fips140

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

// fipsSafeCurves lists only FIPS 140-approved NIST curves, excluding X25519
// which is blocked when Go runs with GODEBUG=fips140=only (Go 1.24+).
var fipsSafeCurves = []tls.CurveID{tls.CurveP384, tls.CurveP256, tls.CurveP521}

// New builds a new client with defaults
func New(config Config) *RestClient {
	if config.EndpointURL == "" {
		config.EndpointURL = DefaultEndpointURL
	}

	client := &RestClient{config: config}

	// set http client timeout
	client.httpClient.Timeout = 15 * time.Second

	// In FIPS 140-only mode (GODEBUG=fips140=only, Go 1.24+) X25519 is
	// blocked. Always set a transport with FIPS-approved NIST curves so that
	// the TLS ClientHello never proposes an unapproved key-share.
	tlsCfg := config.TLSConfig
	if tlsCfg == nil {
		tlsCfg = &tls.Config{}
	}
	tlsCfg.CurvePreferences = fipsSafeCurves
	transport := new(http.Transport)
	transport.TLSClientConfig = tlsCfg
	client.httpClient.Transport = transport

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

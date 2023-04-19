package inmemory

import (
	"time"

	v2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/cli/client/config"
)

// Config describes details associated with making requests
type Config struct {
	url       string
	format    string
	namespace string
	timeout   time.Duration
	tokens    *v2.Tokens
}

// New returns new instance of a config
func New(url string) *Config {
	config := Config{
		url:       url,
		format:    config.FormatJSON,
		namespace: config.DefaultNamespace,
	}

	return &config
}

// APIUrl describes the URL where the API can be found
func (c *Config) APIUrl() string {
	return c.url
}

// Format describes the expected output from the client
func (c *Config) Format() string {
	return c.format
}

// Namespace describes the context of the request
func (c *Config) Namespace() string {
	return c.namespace
}

// Timeout describes the timeout for communicating with the backend
func (c *Config) Timeout() time.Duration {
	return c.timeout
}

// Tokens describes the authorization tokens used to make requests
func (c *Config) Tokens() *v2.Tokens {
	return c.tokens
}

// SaveAPIUrl updates the current value
func (c *Config) SaveAPIUrl(val string) error {
	c.url = val
	return nil
}

// SaveFormat updates the current value
func (c *Config) SaveFormat(val string) error {
	c.format = val
	return nil
}

// SaveNamespace updates the current value
func (c *Config) SaveNamespace(val string) error {
	c.namespace = val
	return nil
}

// SaveTimeout updates the current timeout value
func (c *Config) SaveTimeout(val time.Duration) error {
	c.timeout = val
	return nil
}

// SaveTokens updates the current value
func (c *Config) SaveTokens(val *v2.Tokens) error {
	c.tokens = val
	return nil
}

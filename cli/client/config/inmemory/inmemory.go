package inmemory

import (
	"github.com/sensu/sensu-go/cli/client/config"
	"github.com/sensu/sensu-go/types"
)

// Config describes details associated with making requests
type Config struct {
	url       string
	cafile    string
	edition   string
	format    string
	namespace string
	tokens    *types.Tokens
}

// New returns new instance of a config
func New(url string) *Config {
	config := Config{
		url:       url,
		cafile:    "",
		edition:   config.DefaultEdition,
		format:    config.FormatJSON,
		namespace: config.DefaultNamespace,
	}

	return &config
}

// APIUrl describes the URL where the API can be found
func (c *Config) APIUrl() string {
	return c.url
}

// CAFile describes the location of Trusted CA File
func (c *Config) CAFile() string {
	return c.cafile
}

// Edition describes the edition of the Sensu product
func (c *Config) Edition() string {
	return c.edition
}

// Format describes the expected output from the client
func (c *Config) Format() string {
	return c.format
}

// Namespace describes the context of the request
func (c *Config) Namespace() string {
	return c.namespace
}

// Tokens describes the authorization tokens used to make requests
func (c *Config) Tokens() *types.Tokens {
	return c.tokens
}

// SaveAPIUrl updates the current value
func (c *Config) SaveAPIUrl(val string) error {
	c.url = val
	return nil
}

// SaveCAFile updates the current value
func (c *Config) SaveCAFile(val string) error {
	c.cafile = val
	return nil
}

// SaveEdition updates the current value
func (c *Config) SaveEdition(val string) error {
	c.edition = val
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

// SaveTokens updates the current value
func (c *Config) SaveTokens(val *types.Tokens) error {
	c.tokens = val
	return nil
}

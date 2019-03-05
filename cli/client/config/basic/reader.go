package basic

import (
	"github.com/sensu/sensu-go/cli/client/config"
	"github.com/sensu/sensu-go/types"
)

// APIUrl returns the active cluster API URL
func (c *Config) APIUrl() string {
	return c.Cluster.APIUrl
}

// Format returns the user's preferred format
func (c *Config) Format() string {
	if c.Profile.Format == "" {
		return config.DefaultFormat
	}
	return c.Profile.Format
}

// Namespace returns the user's active namespace
func (c *Config) Namespace() string {
	if c.Profile.Namespace == "" {
		return config.DefaultNamespace
	}
	return c.Profile.Namespace
}

// Tokens returns the active cluster JWT
func (c *Config) Tokens() *types.Tokens {
	return c.Cluster.Tokens
}

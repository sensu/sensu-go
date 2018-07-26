package basic

import (
	"github.com/sensu/sensu-go/cli/client/config"
	"github.com/sensu/sensu-go/types"
)

// APIUrl returns the active cluster API URL
func (c *Config) APIUrl() string {
	return c.Cluster.APIUrl
}

// Edition returns the active cluster edition. Defaults to core
func (c *Config) Edition() string {
	if c.Cluster.Edition == "" {
		return config.DefaultEdition
	}
	return c.Cluster.Edition
}

// Environment returns the user's active environment
func (c *Config) Environment() string {
	if c.Profile.Environment == "" {
		return config.DefaultEnvironment
	}
	return c.Profile.Environment
}

// Format returns the user's preferred format
func (c *Config) Format() string {
	if c.Profile.Format == "" {
		return config.DefaultFormat
	}
	return c.Profile.Format
}

// Organization returns the user's active organization
func (c *Config) Organization() string {
	if c.Profile.Organization == "" {
		return config.DefaultOrganization
	}
	return c.Profile.Organization
}

// Tokens returns the active cluster JWT
func (c *Config) Tokens() *types.Tokens {
	return c.Cluster.Tokens
}

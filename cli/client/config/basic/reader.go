package basic

import (
	"github.com/sensu/sensu-go/types"
)

const (
	defaultEnvironment  = "default"
	defaultFormat       = "none"
	defaultOrganization = "default"
)

// APIUrl returns the active cluster API URL
func (c *Config) APIUrl() string {
	return c.Cluster.APIUrl
}

// Environment returns the user's active environment
func (c *Config) Environment() string {
	if c.Profile.Environment == "" {
		return defaultEnvironment
	}
	return c.Profile.Environment
}

// Format returns the user's preferred format
func (c *Config) Format() string {
	if c.Profile.Format == "" {
		return defaultFormat
	}
	return c.Profile.Format
}

// Organization returns the user's active organization
func (c *Config) Organization() string {
	if c.Profile.Organization == "" {
		return defaultOrganization
	}
	return c.Profile.Organization
}

// Tokens returns the active cluster JWT
func (c *Config) Tokens() *types.Tokens {
	return c.Cluster.Tokens
}

package basic

import (
	"time"

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

// InsecureSkipTLSVerify returns whether InsecureSkipTLSVerify is enabled
func (c *Config) InsecureSkipTLSVerify() bool {
	return c.Cluster.InsecureSkipTLSVerify
}

// Namespace returns the user's active namespace
func (c *Config) Namespace() string {
	if c.Profile.Namespace == "" {
		return config.DefaultNamespace
	}
	return c.Profile.Namespace
}

// Timeout returns the configured timeout
func (c *Config) Timeout() time.Duration {
	if c.Cluster.Timeout == 0*time.Second {
		return config.DefaultTimeout
	}
	return c.Cluster.Timeout
}

// Tokens returns the active cluster JWT
func (c *Config) Tokens() *types.Tokens {
	return c.Cluster.Tokens
}

// TrustedCAFile returns the trusted CA file
func (c *Config) TrustedCAFile() string {
	return c.Cluster.TrustedCAFile
}

package basic

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/sensu/sensu-go/types"
)

// SaveAPIUrl saves the API URL into a configuration file
func (c *Config) SaveAPIUrl(url string) error {
	c.Cluster.APIUrl = url

	return write(c.Cluster, filepath.Join(c.path, clusterFilename))
}

// SaveFormat saves the user's format preference into a configuration file
func (c *Config) SaveFormat(format string) error {
	c.Profile.Format = format

	return write(c.Profile, filepath.Join(c.path, profileFilename))
}

// SaveInsecureSkipTLSVerify saves the InsecureSkipTLSVerify preference
func (c *Config) SaveInsecureSkipTLSVerify(verify bool) error {
	c.Cluster.InsecureSkipTLSVerify = verify

	return write(c.Cluster, filepath.Join(c.path, clusterFilename))
}

// SaveNamespace saves the user's default namespace to a configuration file
func (c *Config) SaveNamespace(namespace string) error {
	c.Profile.Namespace = namespace

	return write(c.Profile, filepath.Join(c.path, profileFilename))
}

// SaveTimeout saves the user's timeout to a configuration file
func (c *Config) SaveTimeout(timeout time.Duration) error {
	c.Cluster.Timeout = timeout

	return write(c.Cluster, filepath.Join(c.path, clusterFilename))
}

// SaveTokens saves the JWT into a configuration file
func (c *Config) SaveTokens(tokens *types.Tokens) error {
	// Update the configuration loaded in memory
	c.Cluster.Tokens = tokens

	// Load the configuration from the file so we don't save any configuration
	// that was overrided with a configuration flag
	savedConfig := &Config{}
	_ = savedConfig.open(filepath.Join(c.path, clusterFilename))
	savedConfig.Cluster.Tokens = tokens

	return write(savedConfig.Cluster, filepath.Join(c.path, clusterFilename))
}

// SaveTrustedCAFile saves the Trusted CA file
func (c *Config) SaveTrustedCAFile(file string) error {
	if file != "" {
		absolute, err := filepath.Abs(file)
		if err != nil {
			return err
		}
		c.Cluster.TrustedCAFile = absolute
	} else {
		c.Cluster.TrustedCAFile = ""
	}

	return write(c.Cluster, filepath.Join(c.path, clusterFilename))
}

func write(data interface{}, path string) error {
	// Make sure the directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, bytes, 0644)
}

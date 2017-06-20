package basic

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

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

// SaveOrganization saves the user's default organization to a configuration file
func (c *Config) SaveOrganization(org string) error {
	c.Profile.Organization = org

	return write(c.Profile, filepath.Join(c.path, profileFilename))
}

// SaveTokens saves the JWT into a configuration file
func (c *Config) SaveTokens(tokens *types.Tokens) error {
	// Update the configuration loaded in memory
	c.Cluster.Tokens = tokens

	// Load the configuration from the file so we don't save any configuration
	// that was overrided with a configuration flag
	savedConfig := &Config{}
	savedConfig.open(filepath.Join(c.path, clusterFilename))
	savedConfig.Cluster.Tokens = tokens

	return write(savedConfig.Cluster, filepath.Join(c.path, clusterFilename))
}

func write(data interface{}, path string) error {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, bytes, 0644)
}

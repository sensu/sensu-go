package config

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"time"

	toml "github.com/pelletier/go-toml"
)

// WriteCredentials writes the given credentials to a file
func (c *MultiConfig) WriteCredentials(url string, token *AccessToken) error {
	config := emptyTomlTree()

	// Read configuration file
	if _, err := os.Stat(CredentialsFilePath); err == nil {
		config, err = toml.LoadFile(CredentialsFilePath)
		if err != nil {
			return fmt.Errorf("Error loading config: %s", err)
		}
	} else {
		// Ensure that the path to the configuration exists
		os.MkdirAll(path.Dir(CredentialsFilePath), 0755)
	}

	// Get the configuation values for the specified profile
	profileVal := c.GetString(profileKey)
	profile, ok := config.Get(profileVal).(*toml.TomlTree)
	if !ok {
		profile = emptyTomlTree()
	}

	// Update profile
	profile.Set("api-url", url)
	profile.Set("secret", token.Token)
	profile.Set("refresh-token", token.RefreshToken)
	profile.Set("expires-at", strconv.FormatInt(token.ExpiresAt.Unix(), 10))
	config.Set(profileKey, profile)

	// Write config
	f, err := os.Create(CredentialsFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = config.WriteTo(f)
	return err
}

func emptyTomlTree() *toml.TomlTree {
	empty, _ := toml.TreeFromMap(make(map[string]interface{}))
	return empty
}

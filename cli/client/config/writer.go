package config

import (
	"fmt"
	"os"
	"path"

	toml "github.com/pelletier/go-toml"
	creds "github.com/sensu/sensu-go/cli/client/credentials"
)

// WriteURL writes the given API URL to the credentials file
func (c *MultiConfig) WriteURL(URL string) error {
	profile := c.GetString(profileKey)
	writer := &credentialsWriter{profile: profile}

	// Read configuration file
	if err := writer.read(); err != nil {
		return err
	}

	// Update profile
	writer.set("api-url", URL)

	// Write config
	err := writer.writeToDisk()
	return err
}

// WriteCredentials writes the given credentials to a file
func (c *MultiConfig) WriteCredentials(token *creds.AccessToken) error {
	profile := c.GetString(profileKey)
	writer := &credentialsWriter{profile: profile}

	// Read configuration file
	if err := writer.read(); err != nil {
		return err
	}

	// Update profile
	writer.set("secret", token.Token)
	writer.set("refresh-token", token.RefreshToken)
	writer.set("expires-at", token.ExpiresAt.Unix())

	// Write config
	err := writer.writeToDisk()
	return err
}

type credentialsWriter struct {
	profile       string
	config        *toml.TomlTree
	relevantCreds *toml.TomlTree
}

func (w *credentialsWriter) read() error {
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

	relevantCreds, ok := config.Get(w.profile).(*toml.TomlTree)
	if !ok {
		relevantCreds = emptyTomlTree()
		config.Set(w.profile, relevantCreds)
	}

	w.config = config
	w.relevantCreds = relevantCreds
	return nil
}

func (w *credentialsWriter) set(key string, data interface{}) {
	w.relevantCreds.Set(key, data)
}

func (w *credentialsWriter) writeToDisk() error {
	// Create or open file in write mode
	f, err := os.Create(CredentialsFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = w.config.WriteTo(f)
	return err
}

func emptyTomlTree() *toml.TomlTree {
	empty, _ := toml.TreeFromMap(make(map[string]interface{}))
	return empty
}

package client

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	assert := assert.New(t)

	// Return includes err when configuration file doesn't exist
	ConfigFilePath = "~/.sensu/config/no_comprende"
	config, err := NewConfig()

	assert.NotNil(config, "NewConfig should still return a valid config")
	assert.NotNil(err, "Error when the file doesn't exist")
}

func TestGet(t *testing.T) {
	assert := assert.New(t)

	// Create new config file for testing
	tmpFile, _ := ioutil.TempFile("", "sensu")
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString(`
[default]
  url = "http://lol.com"
[annastomatoes]
  url = "http://anna.tomato"
	`)
	tmpFile.Close()
	ConfigFilePath = tmpFile.Name()

	config, _ := NewConfig()
	var result string

	// Pull url value from the default profile
	result = config.GetString("url")
	assert.Equal("http://lol.com", result, "correct URL is returned for default profile")

	// Pull url value from annatomatoes profile
	os.Setenv("SENSU_PROFILE", "annastomatoes")
	result = config.GetString("url")
	assert.Equal("http://anna.tomato", result, "correct URL is returned for given profile")

	// Return given (unnested) environment var
	os.Setenv("SENSU_URL", "http://test.local")
	result = config.GetString("url")
	assert.Equal(result, "http://test.local", "ENV variables value is returned")
}

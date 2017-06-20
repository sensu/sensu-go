package basic

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestSaveAPIUrl(t *testing.T) {
	dir, _ := ioutil.TempDir("", "sensu")
	defer os.RemoveAll(dir)

	// Set flags
	flags := pflag.NewFlagSet("config-dir", pflag.ContinueOnError)
	flags.String("config-dir", dir, "")

	config := Load(flags)

	url := "http://127.0.0.1:8080"
	config.SaveAPIUrl(url)
	assert.Equal(t, url, config.APIUrl())
}

func TestSaveFormat(t *testing.T) {
	dir, _ := ioutil.TempDir("", "sensu")
	defer os.RemoveAll(dir)

	// Set flags
	flags := pflag.NewFlagSet("config-dir", pflag.ContinueOnError)
	flags.String("config-dir", dir, "")

	config := Load(flags)

	format := "json"
	config.SaveFormat(format)
	assert.Equal(t, format, config.Format())
}

func TestSaveOrganization(t *testing.T) {
	dir, _ := ioutil.TempDir("", "sensu")
	defer os.RemoveAll(dir)

	// Set flags
	flags := pflag.NewFlagSet("config-dir", pflag.ContinueOnError)
	flags.String("config-dir", dir, "")

	config := Load(flags)

	org := "json"
	config.SaveOrganization(org)
	assert.Equal(t, org, config.Organization())
}

func TestSaveTokens(t *testing.T) {
	dir, _ := ioutil.TempDir("", "sensu")
	defer os.RemoveAll(dir)

	// Set flags
	flags := pflag.NewFlagSet("config-dir", pflag.ContinueOnError)
	flags.String("config-dir", dir, "")

	config := Load(flags)

	tokens := &types.Tokens{Access: "foo"}
	config.SaveTokens(tokens)
	assert.Equal(t, tokens, config.Tokens())
}

func TestSaveTokensWithAPIUrlFlag(t *testing.T) {
	// In case the API URL is passed with a flag, we don't want to save it
	dir, _ := ioutil.TempDir("", "sensu")
	defer os.RemoveAll(dir)

	// Set flags
	flags := pflag.NewFlagSet("api-url", pflag.ContinueOnError)
	flags.String("api-url", "setFromFlag", "")

	dirFlag := pflag.NewFlagSet("config-dir", pflag.ContinueOnError)
	dirFlag.String("config-dir", dir, "")
	flags.AddFlagSet(dirFlag)

	// Create a dummy cluster file
	cluster := &Cluster{APIUrl: "setFromFile"}
	clusterBytes, _ := json.Marshal(cluster)
	clusterPath := filepath.Join(dir, clusterFilename)
	_ = ioutil.WriteFile(clusterPath, clusterBytes, 0644)

	config := Load(flags)

	tokens := &types.Tokens{Access: "foo"}
	config.SaveTokens(tokens)
	assert.Equal(t, tokens, config.Tokens())

	// Make sure we didn't override the orginal API URL
	configFile := Load(dirFlag)
	assert.Equal(t, "setFromFile", configFile.APIUrl())
}

func TestWrite(t *testing.T) {
	dir, _ := ioutil.TempDir("", "sensu")
	defer os.RemoveAll(dir)

	// Set flags
	flags := pflag.NewFlagSet("config-dir", pflag.ContinueOnError)
	flags.String("config-dir", dir, "")

	config := Load(flags)

	url := "http://127.0.0.1:8080"
	config.SaveAPIUrl(url)
	assert.Equal(t, url, config.APIUrl())

	// Reload the config files to make sure the changes were saved
	config2 := Load(flags)
	assert.Equal(t, config.APIUrl(), config2.APIUrl())
}

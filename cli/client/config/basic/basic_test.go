package basic

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlags(t *testing.T) {
	flags := pflag.NewFlagSet("api-url", pflag.ContinueOnError)
	flags.String("api-url", "foo", "")

	config := Load(flags)
	assert.NotNil(t, config)

	assert.Equal(t, "foo", config.APIUrl())
}

func TestLoad(t *testing.T) {
	// Create a dummy directory for testing
	dir, _ := ioutil.TempDir("", "sensu")
	defer func() {
		require.NoError(t, os.RemoveAll(dir))
	}()

	// Set flags
	flags := pflag.NewFlagSet("config-dir", pflag.ContinueOnError)
	flags.String("config-dir", dir, "")

	// Create a dummy cluster file
	cluster := &Cluster{APIUrl: "localhost"}
	clusterBytes, _ := json.Marshal(cluster)
	clusterPath := filepath.Join(dir, clusterFilename)
	_ = ioutil.WriteFile(clusterPath, clusterBytes, 0644)

	// Create a dummy profile file
	profile := &Profile{Format: "json"}
	profileBytes, _ := json.Marshal(profile)
	profilePath := filepath.Join(dir, profileFilename)
	_ = ioutil.WriteFile(profilePath, profileBytes, 0644)

	config := Load(flags)
	assert.NotNil(t, config)
	assert.Equal(t, profile.Format, config.Format())
	assert.Equal(t, cluster.APIUrl, config.APIUrl())
}

func TestLoadMissingFiles(t *testing.T) {
	// Set flags
	flags := pflag.NewFlagSet("config-dir", pflag.ContinueOnError)
	flags.String("config-dir", "/tmp/sensu", "")

	config := Load(flags)
	assert.NotNil(t, config)
}

func TestOpen(t *testing.T) {
	// Create a dummy directory for testing
	dir, err := ioutil.TempDir("", "sensu")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.RemoveAll(dir))
	}()

	// Create a dummy cluster file
	cluster := &Cluster{APIUrl: "localhost"}
	clusterBytes, _ := json.Marshal(cluster)
	clusterPath := filepath.Join(dir, clusterFilename)
	_ = ioutil.WriteFile(clusterPath, clusterBytes, 0644)

	config := &Config{}
	assert.NoError(t, config.open(clusterPath))
	assert.NotEmpty(t, config.APIUrl)
}

func TestOpenMissingFile(t *testing.T) {
	config := &Config{}

	err := config.open("/tmp/sensu/missingfile")
	assert.Error(t, err)
}

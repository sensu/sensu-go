package basic

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestFlags(t *testing.T) {
	flags := pflag.NewFlagSet("api-url", pflag.ContinueOnError)
	flags.String("api-url", "foo", "")
	flags.String("config-dir", t.TempDir(), "")
	v := viper.New()
	_ = v.BindPFlags(flags)

	config := Load(flags, v)
	assert.NotNil(t, config)

	assert.Equal(t, "foo", config.APIUrl())
}

func TestEnv(t *testing.T) {
	flags := pflag.NewFlagSet("api-url", pflag.ContinueOnError)
	_ = os.Setenv("SENSU_API_URL", "foo_env")
	_ = os.Setenv("SENSU_CONFIG_DIR", t.TempDir())
	v, _ := helpers.InitViper(flags)

	config := Load(flags, v)
	assert.NotNil(t, config)

	assert.Equal(t, "foo_env", config.APIUrl())
}

func TestLoad(t *testing.T) {
	// Create a dummy directory for testing
	dir := t.TempDir()

	// Set flags
	flags := pflag.NewFlagSet("config-dir", pflag.ContinueOnError)
	flags.String("config-dir", dir, "")
	v := viper.New()
	_ = v.BindPFlags(flags)

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

	config := Load(flags, v)
	assert.NotNil(t, config)
	assert.Equal(t, profile.Format, config.Format())
	assert.Equal(t, cluster.APIUrl, config.APIUrl())
}

func TestLoadMissingFiles(t *testing.T) {
	// Create a dummy directory for testing
	dir := t.TempDir()

	// Set flags
	flags := pflag.NewFlagSet("config-dir", pflag.ContinueOnError)
	flags.String("config-dir", dir, "")
	v := viper.New()
	_ = v.BindPFlags(flags)

	config := Load(flags, v)
	assert.NotNil(t, config)
}

func TestOpen(t *testing.T) {
	// Create a dummy directory for testing
	dir := t.TempDir()

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
	// Create a dummy directory for testing
	dir := t.TempDir()

	config := &Config{}

	err := config.open(filepath.Join(dir, "missingfile"))
	assert.Error(t, err)
}

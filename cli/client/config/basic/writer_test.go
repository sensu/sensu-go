package basic

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sensu/sensu-go/types"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func tmpDir(t *testing.T) (string, func()) {
	t.Helper()

	dir, err := ioutil.TempDir("", "sensu")
	if err != nil {
		t.Fatal(err)
	}
	cleanup := func() {
		err := os.RemoveAll(dir)
		if err != nil {
			t.Fatal(err)
		}
	}

	return dir, cleanup
}

func TestSaveAPIUrl(t *testing.T) {
	dir, cleanup := tmpDir(t)
	defer cleanup()

	// Set flags
	flags := pflag.NewFlagSet("config-dir", pflag.ContinueOnError)
	flags.String("config-dir", dir, "")

	config := Load(flags)

	url := "http://127.0.0.1:8080"
	require.NoError(t, config.SaveAPIUrl(url))
	assert.Equal(t, url, config.APIUrl())
}

func TestSaveFormat(t *testing.T) {
	dir, cleanup := tmpDir(t)
	defer cleanup()

	// Set flags
	flags := pflag.NewFlagSet("config-dir", pflag.ContinueOnError)
	flags.String("config-dir", dir, "")

	config := Load(flags)

	format := "json"
	require.NoError(t, config.SaveFormat(format))
	assert.Equal(t, format, config.Format())
}

func TestSaveNamespace(t *testing.T) {
	dir, cleanup := tmpDir(t)
	defer cleanup()

	// Set flags
	flags := pflag.NewFlagSet("config-dir", pflag.ContinueOnError)
	flags.String("config-dir", dir, "")

	config := Load(flags)

	namespace := "json"
	require.NoError(t, config.SaveNamespace(namespace))
	assert.Equal(t, namespace, config.Namespace())
}

func TestSaveTimeout(t *testing.T) {
	dir, cleanup := tmpDir(t)
	defer cleanup()

	// Set flags
	flags := pflag.NewFlagSet("config-dir", pflag.ContinueOnError)
	flags.String("config-dir", dir, "")

	config := Load(flags)

	timeout := 30 * time.Second
	require.NoError(t, config.SaveTimeout(timeout))
	assert.Equal(t, timeout, config.Timeout())
}

func TestSaveTokens(t *testing.T) {
	dir, cleanup := tmpDir(t)
	defer cleanup()

	// Set flags
	flags := pflag.NewFlagSet("config-dir", pflag.ContinueOnError)
	flags.String("config-dir", dir, "")

	config := Load(flags)

	tokens := &types.Tokens{Access: "foo"}
	_ = config.SaveTokens(tokens)
	assert.Equal(t, tokens, config.Tokens())
}

func TestSaveTokensWithAPIUrlFlag(t *testing.T) {
	// In case the API URL is passed with a flag, we don't want to save it
	dir, cleanup := tmpDir(t)
	defer cleanup()

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
	require.NoError(t, ioutil.WriteFile(clusterPath, clusterBytes, 0644))

	config := Load(flags)

	tokens := &types.Tokens{Access: "foo"}
	require.NoError(t, config.SaveTokens(tokens))
	assert.Equal(t, tokens, config.Tokens())

	// Make sure we didn't override the orginal API URL
	configFile := Load(dirFlag)
	assert.Equal(t, "setFromFile", configFile.APIUrl())
}

func TestWrite(t *testing.T) {
	dir, cleanup := tmpDir(t)
	defer cleanup()

	// Set flags
	flags := pflag.NewFlagSet("config-dir", pflag.ContinueOnError)
	flags.String("config-dir", dir, "")

	config := Load(flags)

	url := "http://127.0.0.1:8080"
	require.NoError(t, config.SaveAPIUrl(url))
	assert.Equal(t, url, config.APIUrl())

	// Reload the config files to make sure the changes were saved
	config2 := Load(flags)
	assert.Equal(t, config.APIUrl(), config2.APIUrl())
}

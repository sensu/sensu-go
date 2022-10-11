package basic

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"

	corev2 "github.com/sensu/core/v2"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveAPIUrl(t *testing.T) {
	// Create a dummy directory for testing
	dir := t.TempDir()

	// Set flags
	flags := pflag.NewFlagSet("config-dir", pflag.ContinueOnError)
	flags.String("config-dir", dir, "")
	v := viper.New()
	_ = v.BindPFlags(flags)

	config := Load(flags, v)

	url := "http://127.0.0.1:8080"
	require.NoError(t, config.SaveAPIUrl(url))
	assert.Equal(t, url, config.APIUrl())
}

func TestSaveFormat(t *testing.T) {
	// Create a dummy directory for testing
	dir := t.TempDir()

	// Set flags
	flags := pflag.NewFlagSet("config-dir", pflag.ContinueOnError)
	flags.String("config-dir", dir, "")
	v := viper.New()
	_ = v.BindPFlags(flags)

	config := Load(flags, v)

	format := "json"
	require.NoError(t, config.SaveFormat(format))
	assert.Equal(t, format, config.Format())
}

func TestSaveNamespace(t *testing.T) {
	// Create a dummy directory for testing
	dir := t.TempDir()

	// Set flags
	flags := pflag.NewFlagSet("config-dir", pflag.ContinueOnError)
	flags.String("config-dir", dir, "")
	v := viper.New()
	_ = v.BindPFlags(flags)

	config := Load(flags, v)

	namespace := "json"
	require.NoError(t, config.SaveNamespace(namespace))
	assert.Equal(t, namespace, config.Namespace())
}

func TestSaveTimeout(t *testing.T) {
	// Create a dummy directory for testing
	dir := t.TempDir()

	// Set flags
	flags := pflag.NewFlagSet("config-dir", pflag.ContinueOnError)
	flags.String("config-dir", dir, "")
	v := viper.New()
	_ = v.BindPFlags(flags)

	config := Load(flags, v)

	timeout := 30 * time.Second
	require.NoError(t, config.SaveTimeout(timeout))
	assert.Equal(t, timeout, config.Timeout())
}

func TestSaveTokens(t *testing.T) {
	// Create a dummy directory for testing
	dir := t.TempDir()

	// Set flags
	flags := pflag.NewFlagSet("config-dir", pflag.ContinueOnError)
	flags.String("config-dir", dir, "")
	v := viper.New()
	_ = v.BindPFlags(flags)

	config := Load(flags, v)

	tokens := &corev2.Tokens{Access: "foo"}
	_ = config.SaveTokens(tokens)
	assert.Equal(t, tokens, config.Tokens())
}

func TestSaveTokensWithAPIUrlFlag(t *testing.T) {
	// In case the API URL is passed with a flag, we don't want to save it
	// Create a dummy directory for testing
	dir := t.TempDir()

	// Set flags
	flags := pflag.NewFlagSet("api-url", pflag.ContinueOnError)
	flags.String("api-url", "setFromFlag", "")
	flags.String("config-dir", dir, "")
	v := viper.New()
	_ = v.BindPFlags(flags)

	dirFlag := pflag.NewFlagSet("config-dir", pflag.ContinueOnError)
	dirFlag.String("config-dir", dir, "")
	flags.AddFlagSet(dirFlag)
	dirViper := viper.New()
	_ = dirViper.BindPFlags(dirFlag)

	// Create a dummy cluster file
	cluster := &Cluster{APIUrl: "setFromFile"}
	clusterBytes, _ := json.Marshal(cluster)
	clusterPath := filepath.Join(dir, clusterFilename)
	require.NoError(t, ioutil.WriteFile(clusterPath, clusterBytes, 0644))

	config := Load(flags, v)

	tokens := &corev2.Tokens{Access: "foo"}
	require.NoError(t, config.SaveTokens(tokens))
	assert.Equal(t, tokens, config.Tokens())

	// Make sure we didn't override the original API URL
	configFile := Load(flags, dirViper)
	assert.Equal(t, "setFromFile", configFile.APIUrl())
}

func TestWrite(t *testing.T) {
	// Create a dummy directory for testing
	dir := t.TempDir()

	// Set flags
	flags := pflag.NewFlagSet("config-dir", pflag.ContinueOnError)
	flags.String("config-dir", dir, "")
	v := viper.New()
	_ = v.BindPFlags(flags)

	config := Load(flags, v)

	url := "http://127.0.0.1:8080"
	require.NoError(t, config.SaveAPIUrl(url))
	assert.Equal(t, url, config.APIUrl())

	// Reload the config files to make sure the changes were saved
	config2 := Load(flags, v)
	assert.Equal(t, config.APIUrl(), config2.APIUrl())
}

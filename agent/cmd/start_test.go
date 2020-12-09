package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func TestNewAgentConfig(t *testing.T) {
	cmd := &cobra.Command{
		Use: "test",
	}
	if err := handleConfig(cmd, []string{}); err != nil {
		t.Fatal("unexpected error while calling handleConfig: ", err)
	}
	_ = cmd.Flags().Set(flagSubscriptions, "dev,ops")

	cfg, err := NewAgentConfig(cmd)
	if err != nil {
		t.Fatal("unexpected error while calling handleConfig: ", err)
	}

	if !reflect.DeepEqual(cfg.Subscriptions, []string{"dev", "ops"}) {
		t.Fatalf("TestNewAgentConfig() subscriptions = %v, want %v", cfg.Subscriptions, `"dev", "ops"`)
	}
}

func tempConfig(t *testing.T, content string) *os.File {
	t.Helper()

	file, err := ioutil.TempFile(os.TempDir(), "sensu-agent-")
	if err != nil {
		t.Fatalf("error creating tmpFile %q: %s", file.Name(), err)
	}

	_, err = file.WriteString(content)
	if err != nil {
		t.Fatalf("could not write to tmpFile %q: %s", file.Name(), err)
	}

	return file
}

func Test_handleConfig_configFile(t *testing.T) {
	cmd := &cobra.Command{
		Use: "test",
	}

	// The default config file should be used as a fallback if no config file was
	// defined
	if err := handleConfig(cmd, []string{}); err != nil {
		t.Fatal("unexpected error while calling handleConfig: ", err)
	}
	if namespace := viper.GetString(flagNamespace); namespace != "default" {
		t.Fatalf("handleConfig() namespace = %s, want %s", namespace, "default")
	}

	// Create a temporary configuration file to be specified via the flag
	configFileForFlag := tempConfig(t, "namespace: ops")
	defer func() {
		_ = configFileForFlag.Close()
		_ = os.Remove(configFileForFlag.Name())
	}()

	// The configuration file specified via the flag should be used
	if err := handleConfig(cmd, []string{fmt.Sprintf("--%s=%s", flagConfigFile, configFileForFlag.Name())}); err != nil {
		t.Fatal("unexpected error while calling handleConfig: ", err)
	}
	if namespace := viper.GetString(flagNamespace); namespace != "ops" {
		t.Fatalf("handleConfig() namespace = %s, want %s", namespace, "ops")
	}

	// Create a temporary configuration file to be specified via the environment
	// variable
	configFileForEnv := tempConfig(t, "namespace: dev")
	defer func() {
		_ = configFileForEnv.Close()
		_ = os.Remove(configFileForEnv.Name())
	}()

	// The configuration file specified via the environment variable should be
	// used
	os.Setenv("SENSU_CONFIG_FILE", configFileForEnv.Name())
	if err := handleConfig(cmd, []string{}); err != nil {
		t.Fatal("unexpected error while calling handleConfig: ", err)
	}
	if namespace := viper.GetString(flagNamespace); namespace != "dev" {
		t.Fatalf("handleConfig() namespace = %s, want %s", namespace, "dev")
	}

	// The flag should have precedence over the environment variable
	if err := handleConfig(cmd, []string{fmt.Sprintf("--%s=%s", flagConfigFile, configFileForFlag.Name())}); err != nil {
		t.Fatal("unexpected error while calling handleConfig: ", err)
	}
	if namespace := viper.GetString(flagNamespace); namespace != "ops" {
		t.Fatalf("handleConfig() namespace = %s, want %s", namespace, "ops")
	}
}

package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

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

func Test_handleConfig(t *testing.T) {
	cmd := &cobra.Command{
		Use: "test",
	}

	// The default config file should be used as a fallback if no config file was
	// defined
	if err := handleConfig(cmd, []string{}, true); err != nil {
		t.Fatal("unexpected error while calling handleConfig: ", err)
	}
	if host := viper.GetString(flagAgentHost); host != "[::]" {
		t.Fatalf("handleConfig() host = %s, want %s", host, "[::]")
	}

	// Create a temporary configuration file to be specified via the flag
	configFileForFlag := tempConfig(t, "agent-host: localhost")
	defer func() {
		_ = configFileForFlag.Close()
		_ = os.Remove(configFileForFlag.Name())
	}()

	// The configuration file specified via the flag should be used, regardless of
	// its order of appearance
	if err := handleConfig(cmd, []string{
		fmt.Sprintf("--%s=%s", flagLogLevel, "warn"),
		fmt.Sprintf("--%s=%s", flagConfigFile, configFileForFlag.Name()),
	}, true); err != nil {
		t.Fatal("unexpected error while calling handleConfig: ", err)
	}
	if host := viper.GetString(flagAgentHost); host != "localhost" {
		t.Fatalf("handleConfig() host = %s, want %s", host, "localhost")
	}

	// Create a temporary configuration file to be specified via the environment
	// variable
	configFileForEnv := tempConfig(t, "agent-host: 127.0.0.1")
	defer func() {
		_ = configFileForEnv.Close()
		_ = os.Remove(configFileForEnv.Name())
	}()

	// The configuration file specified via the environment variable should be
	// used
	os.Setenv("SENSU_BACKEND_CONFIG_FILE", configFileForEnv.Name())
	if err := handleConfig(cmd, []string{}, true); err != nil {
		t.Fatal("unexpected error while calling handleConfig: ", err)
	}
	if host := viper.GetString(flagAgentHost); host != "127.0.0.1" {
		t.Fatalf("handleConfig() host = %s, want %s", host, "127.0.0.1")
	}

	// The flag should have precedence over the environment variable
	if err := handleConfig(cmd, []string{fmt.Sprintf("--%s=%s", flagConfigFile, configFileForFlag.Name())}, true); err != nil {
		t.Fatal("unexpected error while calling handleConfig: ", err)
	}
	if host := viper.GetString(flagAgentHost); host != "localhost" {
		t.Fatalf("handleConfig() host = %s, want %s", host, "localhost")
	}
}

package testing

import (
	"io/ioutil"
	"os"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client"
	clientmock "github.com/sensu/sensu-go/cli/client/testing"
	"github.com/spf13/cobra"
)

// NewMockCLI return SensuCLI instance w/ mocked values
func NewMockCLI() *cli.SensuCli {
	config := &clientmock.MockConfig{}
	client := &clientmock.MockClient{}

	return &cli.SensuCli{
		Client: client,
		Config: config,
	}
}

// SimpleSensuCLI return SensuCLI instance w/ given client & live config
func SimpleSensuCLI(apiClient client.APIClient) *cli.SensuCli {
	config, _ := client.NewConfig()

	return &cli.SensuCli{
		Client: apiClient,
		Config: config,
	}
}

// RunCmd runs your SensuCLI command and returns any output and errors the
// command might have returnerd. Works with commands that have implemented Run
// or RunE hooks.
func RunCmd(cmd *cobra.Command, args []string) (string, error) {
	var err error

	// So that we can caputre output we reassign cmd.output
	reader, writer, _ := os.Pipe()
	cmd.SetOutput(writer)

	// Run given command
	if cmd.Run != nil {
		cmd.Run(cmd, args)
	} else if cmd.RunE != nil {
		err = cmd.RunE(cmd, args)
	}

	// Close the writer so that we do not run into
	// a deadlock while reading
	writer.Close()

	// Store the contents of the reader as a string
	bytes, _ := ioutil.ReadAll(reader)

	return string(bytes), err
}

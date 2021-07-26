package testing

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/cli/client/config/basic"
	clientmock "github.com/sensu/sensu-go/cli/client/testing"
	"github.com/spf13/cobra"
)

func NewCLI() *cli.SensuCli {
	return NewCLIWithValue("json")
}

// NewCLIWithValue returns a SensuCLI instace with mocked values with json format
func NewCLIWithValue(format string) *cli.SensuCli {
	mockCLI := NewMockCLI()
	config := mockCLI.Config.(*clientmock.MockConfig)
	config.On("Format").Return(format)

	return mockCLI
}

// NewMockCLI return SensuCLI instance w/ default mocked values
func NewMockCLI() *cli.SensuCli {
	return NewMockCLIWithValue("default")
}

// NewMockCLIWithValue return SensuCLI instance w/ mocked values
func NewMockCLIWithValue(namespace string) *cli.SensuCli {
	mockConfig := &clientmock.MockConfig{}
	mockClient := &clientmock.MockClient{}

	// Set defaults ...
	mockConfig.On("Namespace").Return(namespace)

	return &cli.SensuCli{
		Client: mockClient,
		Config: mockConfig,
		InFile: os.Stdin,
	}
}

// SimpleSensuCLI return SensuCLI instance w/ given client & live config
func SimpleSensuCLI(apiClient client.APIClient) *cli.SensuCli {
	c := basic.Load(nil, nil)

	return &cli.SensuCli{
		Client: apiClient,
		Config: c,
	}
}

// RunCmd runs your SensuCLI command and returns any output and errors the
// command might have returned. Works with commands that have implemented Run
// or RunE hooks.
func RunCmd(cmd *cobra.Command, args []string) (string, error) {
	var err error

	// So that we can capture output we reassign cmd.output
	tmpFile, err := ioutil.TempFile(os.TempDir(), "sensu-cli-")
	if err != nil {
		log.Panic("Error creating tmpFile: ", tmpFile.Name())
	}

	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

	cmd.SetOutput(tmpFile)

	// Run given command
	if cmd.Run != nil {
		cmd.Run(cmd, args)
	} else if cmd.RunE != nil {
		err = cmd.RunE(cmd, args)
	}

	// Close the file so that we can read from it
	_ = tmpFile.Close()

	// Store the contents of the reader as a string
	bytes, _ := ioutil.ReadFile(tmpFile.Name())

	return string(bytes), err
}

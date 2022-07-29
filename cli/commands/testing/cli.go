package testing

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/cli/client/config/basic"
	clientmock "github.com/sensu/sensu-go/cli/client/testing"
	"github.com/spf13/cobra"
)

// NewCLI returns a SensuCLI instance with mocked values with json format
func NewCLI() *cli.SensuCli {
	cli := NewMockCLI()
	config := cli.Config.(*clientmock.MockConfig)
	config.On("Format").Return("json")

	return cli
}

func WithMockCLI(tb testing.TB, fn func(cli *cli.SensuCli)) {
	tb.Helper()

	config := &clientmock.MockConfig{}
	client := &clientmock.MockClient{}

	// Set defaults ...
	config.On("Namespace").Return("default")

	// Create temporary files for stdin, stdout & stderr to make it easier to
	// interact with io.
	stdin, err := ioutil.TempFile(os.TempDir(), "sensu-cli-")
	if err != nil {
		tb.Fatal("Error creating stdin file: ", stdin.Name())
	}
	defer func() {
		_ = os.Remove(stdin.Name())
	}()

	stdout, err := ioutil.TempFile(os.TempDir(), "sensu-cli-")
	if err != nil {
		tb.Fatal("Error creating stdout file: ", stdout.Name())
	}
	defer func() {
		_ = os.Remove(stdout.Name())
	}()

	stderr, err := ioutil.TempFile(os.TempDir(), "sensu-cli-")
	if err != nil {
		tb.Fatal("Error creating stderr file: ", stderr.Name())
	}
	defer func() {
		_ = os.Remove(stderr.Name())
	}()

	fn(&cli.SensuCli{
		Client:  client,
		Config:  config,
		InFile:  stdin,
		OutFile: stdout,
		ErrFile: stderr,
	})
}

// NewMockCLI return SensuCLI instance w/ mocked values
func NewMockCLI() *cli.SensuCli {
	config := &clientmock.MockConfig{}
	client := &clientmock.MockClient{}

	// Set defaults ...
	config.On("Namespace").Return("default")

	// Create temporary files for stdin, stdout & stderr to make it easier to
	// interact with io.
	stdin, err := ioutil.TempFile(os.TempDir(), "sensu-cli-")
	if err != nil {
		log.Panic("Error creating stdin file: ", stdin.Name())
	}
	defer func() {
		_ = os.Remove(stdin.Name())
	}()

	stdout, err := ioutil.TempFile(os.TempDir(), "sensu-cli-")
	if err != nil {
		log.Panic("Error creating stdout file: ", stdout.Name())
	}
	defer func() {
		_ = os.Remove(stdout.Name())
	}()

	stderr, err := ioutil.TempFile(os.TempDir(), "sensu-cli-")
	if err != nil {
		log.Panic("Error creating stderr file: ", stderr.Name())
	}
	defer func() {
		_ = os.Remove(stderr.Name())
	}()

	return &cli.SensuCli{
		Client:  client,
		Config:  config,
		InFile:  stdin,
		OutFile: stdout,
		ErrFile: stderr,
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

	tmpFile, err := ioutil.TempFile(os.TempDir(), "sensu-cli-")
	if err != nil {
		log.Panic("Error creating tmp file: ", tmpFile.Name())
	}
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

	return RunCmdWithOutFile(cmd, args, tmpFile)
}

func RunCmdWithOutFile(cmd *cobra.Command, args []string, outFile *os.File) (string, error) {
	var err error

	cmd.SetOutput(outFile)

	// Run given command
	if cmd.Run != nil {
		cmd.Run(cmd, args)
	} else if cmd.RunE != nil {
		err = cmd.RunE(cmd, args)
	}

	// Close the file so that we can read from it
	_ = outFile.Close()

	// Store the contents of the reader as a string
	bytes, _ := ioutil.ReadFile(outFile.Name())

	return string(bytes), err
}

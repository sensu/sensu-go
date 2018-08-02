package cli

import (
	"os"

	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/cli/client/config"
	"github.com/sensu/sensu-go/cli/client/config/basic"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

// SensuCmdName name of the command
const SensuCmdName = "sensuctl"

// TypeError is the string returned in the event of an invalid type assertion
const TypeError = "TypeError"

// SensuCli is an instance of the Sensu command line client;
// encapsulates API client, logger & general configuration.
type SensuCli struct {
	Config config.Config
	Client client.APIClient
	Logger *logrus.Entry
	InFile *os.File
}

// New SensuCLI given persistent flags from command
func New(flags *pflag.FlagSet) *SensuCli {
	conf := basic.Load(flags)
	client := client.New(conf)
	logger := logrus.WithFields(logrus.Fields{
		"component": "cli-client",
	})

	return &SensuCli{
		Client: client,
		Config: conf,
		Logger: logger,
		InFile: os.Stdin,
	}
}

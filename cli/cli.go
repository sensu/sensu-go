package cli

import (
	"github.com/Sirupsen/logrus"
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/cli/client/config"
	"github.com/sensu/sensu-go/cli/client/config/basic"
	"github.com/spf13/pflag"
)

const SensuCmdName = "sensuctl"

// SensuCli is an instance of the Sensu command line client;
// encapsulates API client, logger & general configuration.
type SensuCli struct {
	Config config.Config
	Client client.APIClient
	Logger *logrus.Entry
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
	}
}

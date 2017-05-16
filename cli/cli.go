package cli

import (
	"github.com/Sirupsen/logrus"
	"github.com/sensu/sensu-go/cli/client"
	"github.com/spf13/pflag"
)

// SensuCli is an instance of the Sensu command line client;
// encapsulates API client, logger & general configuration.
type SensuCli struct {
	Config client.Config
	Client client.APIClient
	Logger *logrus.Entry
}

// New SensuCLI given persistent flags from command
func New(flags *pflag.FlagSet) *SensuCli {
	clientConfig, _ := client.NewConfig()
	clientConfig.BindPFlag("api-url", flags.Lookup("api-url"))
	clientConfig.BindPFlag("secret", flags.Lookup("api-secret"))
	clientConfig.BindPFlag("profile", flags.Lookup("profile"))

	client := client.New(clientConfig)
	logger := logrus.WithFields(logrus.Fields{
		"component": "cli-client",
	})

	return &SensuCli{
		Client: client,
		Config: clientConfig,
		Logger: logger,
	}
}

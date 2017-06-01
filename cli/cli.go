package cli

import (
	"github.com/Sirupsen/logrus"
	"github.com/sensu/sensu-go/cli/client"
	clientconfig "github.com/sensu/sensu-go/cli/client/config"
	"github.com/spf13/pflag"
)

// SensuCli is an instance of the Sensu command line client;
// encapsulates API client, logger & general configuration.
type SensuCli struct {
	Config clientconfig.Config
	Client client.APIClient
	Logger *logrus.Entry
}

// New SensuCLI given persistent flags from command
func New(flags *pflag.FlagSet) *SensuCli {
	clientConfig, _ := clientconfig.NewConfig()
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

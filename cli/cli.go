package cli

import (
	"path/filepath"

	"github.com/Sirupsen/logrus"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/cli/client/config"
	"github.com/sensu/sensu-go/cli/client/config/basic"
	"github.com/spf13/pflag"
)

var (
	// sensuPath contains path to CLI configuration files
	sensuPath string
)

func init() {
	h, _ := homedir.Dir()
	sensuPath = filepath.Join(h, ".config", "sensu")
}

// SensuCli is an instance of the Sensu command line client;
// encapsulates API client, logger & general configuration.
type SensuCli struct {
	Config config.Config
	Client client.APIClient
	Logger *logrus.Entry
}

// New SensuCLI given persistent flags from command
func New(flags *pflag.FlagSet) *SensuCli {
	conf := basic.Load(flags, sensuPath)

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

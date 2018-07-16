package basic

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/types"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

const (
	clusterFilename = "cluster"
	profileFilename = "profile"
)

var logger = logrus.WithFields(logrus.Fields{
	"component": "cli-config",
})

// Config contains the CLI configuration
type Config struct {
	Cluster
	Profile
	path string
}

// Cluster contains the Sensu cluster access information
type Cluster struct {
	APIUrl  string `json:"api-url"`
	Edition string `json:"edition"`
	*types.Tokens
}

// Profile contains the active configuration
type Profile struct {
	Environment  string `json:"environment"`
	Format       string `json:"format"`
	Organization string `json:"organization"`
}

// Load imports the CLI configuration and returns an initialized Config struct
func Load(flags *pflag.FlagSet) *Config {
	conf := &Config{}

	// Retrieve the path of the configuration directory
	if flags != nil {
		// NOTE:
		//
		// We have a significant order of operations problem where, we need
		// the flags parsed to get the current config file, however, we need the
		// values from the config file to properly set up the flags.
		flags.SetOutput(ioutil.Discard)
		_ = flags.Parse(os.Args[1:])

		if value, err := flags.GetString("config-dir"); err == nil && value != "" {
			conf.path = value
		}
	}

	// Load the profile config file
	if err := conf.open(profileFilename); err != nil {
		logger.Debug(err)
	}

	// Load the cluster config file
	if err := conf.open(clusterFilename); err != nil {
		logger.Debug(err)
	}

	// Override environment
	if flags != nil {
		if value := helpers.GetChangedStringValueFlag("environment", flags); value != "" {
			conf.Profile.Environment = value
		}

		// Override organization
		if value := helpers.GetChangedStringValueFlag("organization", flags); value != "" {
			conf.Profile.Organization = value
		}
	}

	// Load the flags config
	conf.flags(flags)

	return conf
}

func (c *Config) flags(flags *pflag.FlagSet) {
	if flags == nil {
		return
	}

	// Set the API URL
	if value, err := flags.GetString("api-url"); err == nil && value != "" {
		c.Cluster.APIUrl = value
	}
}

func (c *Config) open(path string) error {
	content, err := ioutil.ReadFile(filepath.Join(c.path, path))
	if err != nil {
		return err
	}

	return json.Unmarshal(content, c)
}

package basic

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/pflag"
)

const (
	clusterFilename = "cluster"
	profileFilename = "profile"
)

var logger = logrus.WithFields(logrus.Fields{
	"component": "cli-config",
})

// OrganizationDefault default value to use for organization
const OrganizationDefault = "default"

// EnvironmentDefault default value to use for organization
const EnvironmentDefault = "default"

// Config contains the CLI configuration
type Config struct {
	Cluster
	Profile
	path string
}

// Cluster contains the Sensu cluster access information
type Cluster struct {
	APIUrl string `json:"api-url"`
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
	config := &Config{}

	// Retrieve the path of the configuration directory
	if flags != nil {
		// NOTE:
		//
		// We have a significant order of operations problem where, we need
		// the flags parsed to get the current config file, however, we need the
		// values from the config file to properly set up the flags.
		flags.SetOutput(ioutil.Discard)
		flags.Parse(os.Args[1:])

		if value, err := flags.GetString("config-dir"); err == nil && value != "" {
			config.path = value
		}
	}

	// Load the profile config file
	if err := config.open(profileFilename); err != nil {
		logger.Debug(err)
	}

	// Load the cluster config file
	if err := config.open(clusterFilename); err != nil {
		logger.Debug(err)
	}

	// Override environment
	if flags != nil {
		if value, err := flags.GetString("environment"); err == nil {
			if value != "" {
				config.Profile.Environment = value
			} else if config.Profile.Environment == "" {
				config.Profile.Environment = defaultEnvironment
			}

			if flag := flags.Lookup("environment"); flag != nil {
				flag.DefValue = config.Profile.Environment
			}
		}

		// Override organization
		if value, err := flags.GetString("organization"); err == nil {
			if value != "" {
				config.Profile.Organization = value
			} else if config.Profile.Organization == "" {
				config.Profile.Organization = defaultOrganization
			}

			if flag := flags.Lookup("organization"); flag != nil {
				flag.DefValue = config.Profile.Organization
			}
		}
	}

	// Load the flags config
	config.flags(flags)

	return config
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

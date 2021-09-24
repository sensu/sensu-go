package basic

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/types"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
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
	APIUrl                string `json:"api-url"`
	TrustedCAFile         string `json:"trusted-ca-file"`
	InsecureSkipTLSVerify bool   `json:"insecure-skip-tls-verify"`
	*types.Tokens
	APIKey  string
	Timeout time.Duration `json:"timeout"`
}

// Profile contains the active configuration
type Profile struct {
	Format    string `json:"format"`
	Namespace string `json:"namespace"`
}

// Load imports the CLI configuration and returns an initialized Config struct
func Load(flags *pflag.FlagSet, v *viper.Viper) *Config {
	conf := &Config{}

	// Retrieve the path of the configuration directory
	if flags != nil && v != nil {
		// When Load() is called, some sub-command local flags, such as
		// --format, are not registered yet and this leads to "unknown flags"
		// errors being returned by cobra. Such an error can throw off the cobra
		// parser, leading to all the flags appearing after the offending,
		// supposedly unknown flag to be ignored.
		//
		// For now, we just ignore such "unknown flags" errors in order not to
		// potentially ignore other flags as a side effect. cobra will still store
		// the name/value of all the flags, even if they are registered later.
		//
		// Unfortunately, a (rather big) refactor of the CLI is probably the
		// only way to get around that. Such a refactor would involve completely
		// reordering the way we define commands, load the configuration files
		// and override properties with command line flags.
		flags.ParseErrorsWhitelist = pflag.ParseErrorsWhitelist{UnknownFlags: true}
		_ = flags.Parse(os.Args[1:])

		if value := v.GetString("config-dir"); value != "" {
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

	if v != nil {
		// Override namespace
		if value := helpers.GetChangedStringValueEnv("namespace", v); value != "" {
			conf.Profile.Namespace = value
		}
	}

	// Load the flags config
	conf.flags(v)

	return conf
}

func (c *Config) flags(v *viper.Viper) {
	if v == nil {
		return
	}

	if value := v.GetString("api-url"); value != "" {
		c.Cluster.APIUrl = value
	}
	if value := v.GetString("api-key"); value != "" {
		c.Cluster.APIKey = value
	}
	if value := v.GetBool("insecure-skip-tls-verify"); value {
		c.Cluster.InsecureSkipTLSVerify = value
	}
	if value := v.GetString("trusted-ca-file"); value != "" {
		c.Cluster.TrustedCAFile = value
	}
	if value := v.GetString("timeout"); value != "" {
		duration, err := time.ParseDuration(value)
		if err == nil {
			c.Cluster.Timeout = duration
		} else {
			// Default to timeout of 15 seconds
			c.Cluster.Timeout = 15 * time.Second
		}
	}
}

func (c *Config) open(path string) error {
	content, err := ioutil.ReadFile(filepath.Join(c.path, path))
	if err != nil {
		return err
	}

	return json.Unmarshal(content, c)
}

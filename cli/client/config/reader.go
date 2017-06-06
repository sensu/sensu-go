package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// NewConfig reads configuration file, sets up ENV variables,
// configures defaults and returns new a Config w/ given values.
func NewConfig() (*MultiConfig, error) {
	credentialsConf, err := newCredentialsConfig()
	config := &MultiConfig{
		profile:     newProfilesConfig(),
		credentials: credentialsConf,
	}

	return config, err
}

func newProfilesConfig() *viper.Viper {
	v := viper.New()

	// Set the default profile
	v.SetDefault(profileKey, profileDefault)

	// ENV variables
	v.SetEnvPrefix("SENSU")
	v.BindEnv(profileKey)

	return v
}

func newCredentialsConfig() (*viper.Viper, error) {
	v := viper.New()

	// ENV variables
	v.SetEnvPrefix("SENSU")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.BindEnv("api-url")
	v.BindEnv("secret")

	// Configuration file
	v.SetConfigFile(CredentialsFilePath)
	v.SetConfigType("toml")

	// Open the configuration file and watch it in case the token is refeshed
	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		// NOTE: https://github.com/spf13/viper/pull/299
		// TODO: log something?
	})

	err := v.ReadInConfig()
	return v, err
}

// Get value from configuration for given key
func (c *MultiConfig) Get(key string) interface{} {
	if key == profileKey {
		return c.profile.Get(key)
	} else if val := c.credentials.Get(key); val != "" && val != nil {
		return val
	}

	key = c.prependProfileTo(key)
	return c.credentials.Get(key)
}

// GetString value from configuration for given key
func (c *MultiConfig) GetString(key string) string {
	val, _ := c.Get(key).(string)
	return val
}

// GetTime value from configuration for given key
func (c *MultiConfig) GetTime(key string) time.Time {
	key = c.prependProfileTo(key)
	return c.credentials.GetTime(key)
}

func (c *MultiConfig) prependProfileTo(key string) string {
	return fmt.Sprintf("%s.%s", c.Get(profileKey), key)
}

// BindPFlag binds given pflag to config
func (c *MultiConfig) BindPFlag(key string, flag *pflag.Flag) {
	if flag != nil {
		c.profile.BindPFlag(key, flag)
		c.credentials.BindPFlag(key, flag)
	}
}

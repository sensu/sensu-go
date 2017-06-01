package config

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Config ...
type Config interface {
	Get(key string) interface{}
	GetString(key string) string
	BindPFlag(key string, flag *pflag.Flag)
}

// MultiConfig wraps viper library
type MultiConfig struct {
	// Handles determining the curretly selected profile
	profile *viper.Viper

	// Handler retrieving credentials for the selected profile
	credentials *viper.Viper
}

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
	err := v.ReadInConfig() //
	v.WatchConfig()

	return v, err
}

// Get value from configuration for given key
func (c *MultiConfig) Get(key string) interface{} {
	if key == profileKey {
		return c.profile.Get(key)
	} else if val := c.credentials.Get(key); val != "" && val != nil {
		return val
	}

	key = fmt.Sprintf("%s.%s", c.Get(profileKey), key)
	return c.credentials.Get(key)
}

// GetString value from configuration for given key
func (c *MultiConfig) GetString(key string) string {
	val, _ := c.Get(key).(string)
	return val
}

// BindPFlag binds given pflag to config
func (c *MultiConfig) BindPFlag(key string, flag *pflag.Flag) {
	if flag != nil {
		c.profile.BindPFlag(key, flag)
		c.credentials.BindPFlag(key, flag)
	}
}

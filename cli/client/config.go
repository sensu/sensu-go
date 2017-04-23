package client

import (
	"fmt"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	// SENSU_PROFILE or --profile
	profileKey     = "profile"
	profileDefault = "default"
)

type Config struct {
	viper *viper.Viper
}

// NewConfig reads configuration file, sets up ENV variables,
// configures defaults and returns new a Config w/ given values.
func NewConfig() (*Config, error) {
	v := viper.New()

	// Set the default profile
	v.SetDefault(profileKey, profileDefault)

	// ENV variables
	v.SetEnvPrefix("SENSU")
	v.BindEnv(profileKey)
	v.BindEnv("url")

	// Configuration file
	v.AddConfigPath("$HOME/.config/sensu")
	v.SetConfigName("profiles")
	v.SetConfigType("toml")

	err := v.ReadInConfig()
	return &Config{viper: v}, err
}

func (c *Config) Get(key string) interface{} {
	if val := c.viper.Get(key); val != "" {
		logger.Info("here", val)
		return val
	}

	key = fmt.Sprintf("%s.%s", c.GetString(profileKey), key)
	logger.Info("not here", key, c.viper.Get(key))
	return c.viper.Get(key)
}

func (c *Config) GetString(key string) string {
	val, _ := c.Get(key).(string)
	return val
}

func (c *Config) BindPFlag(key string, flag *pflag.Flag) {
	c.viper.BindPFlag(key, flag)
}

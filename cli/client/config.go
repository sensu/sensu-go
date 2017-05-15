package client

import (
	"fmt"
	"path/filepath"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	// SENSU_PROFILE or --profile
	profileKey     = "profile"
	profileDefault = "default"
)

var (
	// ConfigFilePath contains path to configuration file
	ConfigFilePath string
)

// Config ...
type Config interface {
	Get(key string) interface{}
	GetString(key string) string
	BindPFlag(key string, flag *pflag.Flag)
}

// MultiConfig wraps viper library
type MultiConfig struct {
	viper *viper.Viper
}

// NewConfig reads configuration file, sets up ENV variables,
// configures defaults and returns new a Config w/ given values.
func NewConfig() (*MultiConfig, error) {
	v := viper.New()

	// Set the default profile
	v.SetDefault(profileKey, profileDefault)

	// ENV variables
	v.SetEnvPrefix("SENSU")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.BindEnv(profileKey)
	v.BindEnv("api-url")
	v.BindEnv("secret")

	// Configuration file
	v.SetConfigFile(ConfigFilePath)
	v.SetConfigType("toml")

	err := v.ReadInConfig()
	return &MultiConfig{viper: v}, err
}

// Get value from configuration for given key
func (c *MultiConfig) Get(key string) interface{} {
	if val := c.viper.Get(key); val != "" && val != nil {
		return val
	}

	key = fmt.Sprintf("%s.%s", c.Get(profileKey), key)
	return c.viper.Get(key)
}

// GetString value from configuration for given key
func (c *MultiConfig) GetString(key string) string {
	val, _ := c.Get(key).(string)
	return val
}

// BindPFlag binds given pflag to config
func (c *MultiConfig) BindPFlag(key string, flag *pflag.Flag) {
	if flag != nil {
		c.viper.BindPFlag(key, flag)
	}
}

func init() {
	h, _ := homedir.Dir()
	ConfigFilePath = filepath.Join(h, ".config", "sensu", "profiles")
}

package client

import "github.com/spf13/viper"

func NewConfig() (c *viper.Viper, err error) {
	c = config.New()
	c.SetConfigName("profiles")
	c.SetConfigPath("$HOME/.config/sensu")
	c.SetConfigType("TOML")

	if err = c.ReadInConfig(); err != nil {
		return
	}
}

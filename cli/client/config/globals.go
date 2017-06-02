package config

import (
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

const (
	// SENSU_PROFILE or --profile
	profileKey     = "profile"
	profileDefault = "default"
)

var (
	// CredentialsFilePath contains path to configuration file
	CredentialsFilePath string
)

// MultiConfig wraps viper library
type MultiConfig struct {
	// Handles determining the curretly selected profile
	profile *viper.Viper

	// Handler retrieving credentials for the selected profile
	credentials *viper.Viper
}

func init() {
	h, _ := homedir.Dir()
	CredentialsFilePath = filepath.Join(h, ".config", "sensu", "credentials")
}

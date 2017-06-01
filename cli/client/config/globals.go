package config

import (
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
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

func init() {
	h, _ := homedir.Dir()
	CredentialsFilePath = filepath.Join(h, ".config", "sensu", "credentials")
}

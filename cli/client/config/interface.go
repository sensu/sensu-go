package config

import (
	creds "github.com/sensu/sensu-go/cli/client/credentials"
	"github.com/spf13/pflag"
)

// Config ...
type Config interface {
	ReadConfig
	WriteConfig
}

// ReadConfig ...
type ReadConfig interface {
	Get(key string) interface{}
	GetString(key string) string
	BindPFlag(key string, flag *pflag.Flag)
}

// WriteConfig ...
type WriteConfig interface {
	WriteURL(URL string) error
	WriteCredentials(token *creds.AccessToken) error
}

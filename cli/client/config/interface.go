package config

import (
	"time"

	"github.com/sensu/sensu-go/types"
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
	GetTime(key string) time.Time
	BindPFlag(key string, flag *pflag.Flag)
}

// WriteConfig ...
type WriteConfig interface {
	WriteURL(URL string) error
	WriteCredentials(tokens *types.Tokens) error
}

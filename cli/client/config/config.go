package config

import "github.com/sensu/sensu-go/types"

// Config represents an abstracted configuration
type Config interface {
	Read
	Write
}

// Read contains all methods related to reading configuration
type Read interface {
	APIUrl() string
	Format() string
	Environment() string
	Organization() string
	Tokens() *types.Tokens
}

// Write contains all methods related to setting and writting configuration
type Write interface {
	SaveAPIUrl(string) error
	SaveFormat(string) error
	SaveEnvironment(string) error
	SaveOrganization(string) error
	SaveTokens(*types.Tokens) error
}

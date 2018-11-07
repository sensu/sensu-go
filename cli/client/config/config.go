package config

import (
	"github.com/sensu/sensu-go/types"
)

const (
	// DefaultEdition is the default Sensu edition
	DefaultEdition = types.CoreEdition

	// DefaultNamespace represents the default namespace
	DefaultNamespace = "default"

	// DefaultFormat is the default format output for printers.
	DefaultFormat = FormatTabular

	// FormatTabular indicates tabular format for printers.
	FormatTabular = "tabular"

	// FormatJSON indicates JSON format for printers.
	FormatJSON = "json"

	// FormatWrappedJSON indicates wrapped JSON format for printers.
	FormatWrappedJSON = "wrapped-json"

	// FormatYAML indicates YAML format for printers. It has the same layout
	// as wrapped JSON.
	FormatYAML = "yaml"
)

// Config is an abstract configuration
type Config interface {
	Read
	Write
}

// Read contains all methods related to reading configuration
type Read interface {
	APIUrl() string
	Edition() string
	Format() string
	Namespace() string
	Tokens() *types.Tokens
}

// Write contains all methods related to setting and writting configuration
type Write interface {
	SaveAPIUrl(string) error
	SaveEdition(string) error
	SaveFormat(string) error
	SaveNamespace(string) error
	SaveTokens(*types.Tokens) error
}

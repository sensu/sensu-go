package helpers

import (
	"fmt"
	"io"
	"net/http"

	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/cli/client/config"
	"github.com/sensu/sensu-go/cli/commands/flags"
	"github.com/sensu/sensu-go/cli/elements/list"
	"github.com/spf13/cobra"
)

type printTableFunc func(interface{}, io.Writer)

// HeaderWarning is the header key for entity limit warnings
const HeaderWarning = "Sensu-Entity-Warning"

// PrintList prints a list of resources to stdout with a title, if relevant.
func PrintList(cmd *cobra.Command, format string, printTable printTableFunc, objects []corev3.Resource, v interface{}, header http.Header) error {
	if warning := header.Get(HeaderWarning); warning != "" {
		if err := PrintTitle(GetChangedStringValueViper(flags.Format, cmd.Flags()), format, warning, cmd.OutOrStdout()); err != nil {
			return err
		}
	}
	return Print(cmd, format, printTable, objects, v)
}

// Print displays
func Print(cmd *cobra.Command, format string, printTable printTableFunc, objects []corev3.Resource, v interface{}) error {
	viper, err := InitViper(cmd.Flags())
	if err != nil {
		return err
	}

	if f := GetChangedStringValueEnv(flags.Format, viper); f != "" {
		format = f
	}
	switch format {
	case config.FormatJSON:
		if objects == nil {
			return PrintJSON(v, cmd.OutOrStdout())
		}
		return PrintResourceListJSON(objects, cmd.OutOrStdout())
	case config.FormatYAML:
		if objects == nil {
			return PrintYAML(v, cmd.OutOrStdout())
		}
		return PrintYAML(objects, cmd.OutOrStdout())
	default:
		printTable(v, cmd.OutOrStdout())
	}

	return nil
}

// PrintFormatted prints the provided interface in the specified format.
// flag overrides the cli config format if set
func PrintFormatted(flag string, format string, v interface{}, w io.Writer, printToList func(interface{}, io.Writer) error) error {
	if flag != "" {
		format = flag
	}
	switch format {
	case config.FormatJSON:
		r, ok := v.(corev3.Resource)
		if !ok {
			return fmt.Errorf("%t is not a Resource", v)
		}
		return PrintResourceJSON(r, w)
	case config.FormatYAML:
		return PrintYAML(v, w)
	default:
		return printToList(v, w)
	}
}

// PrintTitle prints a title for tabular format only.
// Flag overrides the cli config format, if set.
func PrintTitle(flag string, format string, title string, w io.Writer) error {
	if flag != "" {
		format = flag
	}
	// checking the formats exclusively to cover invalid formats
	// that get defaulted to tabular
	if format != config.FormatJSON && format != config.FormatYAML {
		cfg := &list.Config{
			Title: title,
		}
		return list.Print(w, cfg)
	}
	return nil
}

package helpers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/sensu/sensu-go/cli/client/config"
	"github.com/sensu/sensu-go/cli/commands/flags"
	"github.com/sensu/sensu-go/cli/elements/list"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

type printTableFunc func(interface{}, io.Writer)

// HeaderWarning is the header key for entity limit warnings
const HeaderWarning = "Sensu-Entity-Warning"

// PrintList prints a list of resources to stdout with a title, if relevant.
func PrintList(cmd *cobra.Command, format string, printTable printTableFunc, objects []types.Resource, v interface{}, header http.Header) error {
	if warning := header.Get(HeaderWarning); warning != "" {
		if err := PrintTitle(GetChangedStringValueFlag(flags.Format, cmd.Flags()), format, warning, cmd.OutOrStdout()); err != nil {
			return err
		}
	}
	return Print(cmd, format, printTable, objects, v)
}

// Print displays
func Print(cmd *cobra.Command, format string, printTable printTableFunc, objects []types.Resource, v interface{}) error {
	if f := GetChangedStringValueFlag(flags.Format, cmd.Flags()); f != "" {
		format = f
	}
	switch format {
	case config.FormatJSON:
		return PrintJSON(v, cmd.OutOrStdout())
	case config.FormatWrappedJSON:
		if objects == nil {
			return PrintJSON(v, cmd.OutOrStdout())
		}
		return PrintWrappedJSONList(objects, cmd.OutOrStdout())
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
		return PrintJSON(v, w)
	case config.FormatWrappedJSON:
		r, ok := v.(types.Resource)
		if !ok {
			return fmt.Errorf("%t is not a Resource", v)
		}
		return PrintWrappedJSON(r, w)
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
	if format != config.FormatJSON && format != config.FormatWrappedJSON && format != config.FormatYAML {
		cfg := &list.Config{
			Title: title,
		}
		return list.Print(w, cfg)
	}
	return nil
}

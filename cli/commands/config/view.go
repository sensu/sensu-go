package config

import (
	"fmt"
	"io"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/commands/hooks"
	"github.com/sensu/sensu-go/cli/elements/list"
	"github.com/spf13/cobra"
)

// ViewCommand ...
func ViewCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "view",
		Short:        "Display active configuration",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			activeConfig := map[string]string{
				"api-url":      cli.Config.APIUrl(),
				"environment":  cli.Config.Environment(),
				"organization": cli.Config.Organization(),
				"format":       cli.Config.Format(),
			}

			// Determine the format to use to output the data
			flag := helpers.GetChangedStringValueFlag("format", cmd.Flags())
			format := cli.Config.Format()
			return helpers.PrintFormatted(flag, format, activeConfig, cmd.OutOrStdout(), printToList)
		},
		Annotations: map[string]string{
			// We want to be able to run this command regardless of whether the CLI
			// has been configured.
			hooks.ConfigurationRequirement: hooks.ConfigurationNotRequired,
		},
	}

	helpers.AddFormatFlag(cmd.Flags())

	return cmd
}

func printToList(v interface{}, writer io.Writer) error {
	r, ok := v.(map[string]string)
	if !ok {
		return fmt.Errorf("%t is not a map of strings", v)
	}

	cfg := &list.Config{
		Title: "Active Configuration",
		Rows: []*list.Row{
			{
				Label: "API URL",
				Value: r["api-url"],
			},
			{
				Label: "Environment",
				Value: r["environment"],
			},
			{
				Label: "Organization",
				Value: r["organization"],
			},
			{
				Label: "Format",
				Value: r["format"],
			},
		},
	}

	return list.Print(writer, cfg)
}

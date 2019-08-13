package config

import (
	"errors"
	"fmt"
	"io"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/commands/hooks"
	"github.com/sensu/sensu-go/cli/elements/list"
	"github.com/spf13/cobra"
)

// ViewCommand defines subcommand to view active configuration
func ViewCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "view",
		Short:        "Display active configuration",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			username := helpers.GetCurrentUsername(cli.Config)
			if username == "" {
				return errors.New("no active configuration found")
			}
			activeConfig := map[string]string{
				"api-url":   cli.Config.APIUrl(),
				"namespace": cli.Config.Namespace(),
				"format":    cli.Config.Format(),
				"username":  helpers.GetCurrentUsername(cli.Config),
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
				Label: "Namespace",
				Value: r["namespace"],
			},
			{
				Label: "Format",
				Value: r["format"],
			},
			{
				Label: "Username",
				Value: r["username"],
			},
		},
	}

	return list.Print(writer, cfg)
}

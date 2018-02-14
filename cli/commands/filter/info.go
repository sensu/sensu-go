package filter

import (
	"errors"
	"io"
	"strings"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/list"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// InfoCommand defines the 'filter info' subcommand
func InfoCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "info [NAME]",
		Short:        "show detailed filter information",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			format, _ := cmd.Flags().GetString("format")

			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			// Fetch the filter from API
			name := args[0]
			r, err := cli.Client.FetchFilter(name)
			if err != nil {
				return err
			}

			if format == "json" {
				if err := helpers.PrintJSON(r, cmd.OutOrStdout()); err != nil {
					return err
				}
			} else {
				printToList(r, cmd.OutOrStdout())
			}

			return nil
		},
	}

	helpers.AddFormatFlag(cmd.Flags())

	return cmd
}

func printToList(filter *types.EventFilter, writer io.Writer) {
	cfg := &list.Config{
		Title: filter.Name,
		Rows: []*list.Row{
			{
				Label: "Name",
				Value: filter.Name,
			},
			{
				Label: "Action",
				Value: filter.Action,
			},
			{
				Label: "Statements",
				Value: strings.Join(filter.Statements, " && "),
			},
			{
				Label: "Organization",
				Value: filter.Organization,
			},
			{
				Label: "Environment",
				Value: filter.Environment,
			},
		},
	}

	list.Print(writer, cfg)
}

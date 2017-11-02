package filter

import (
	"io"
	"strings"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/flags"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/table"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// ListCommand defines the 'filter list' subcommand
func ListCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "list",
		Short:        "list filters",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			org := cli.Config.Organization()
			if ok, _ := cmd.Flags().GetBool(flags.AllOrgs); ok {
				org = "*"
			}

			// Fetch filters from the API
			r, err := cli.Client.ListFilters(org)
			if err != nil {
				return err
			}

			// Determine the format to use to output the data
			var format string
			if format, _ = cmd.Flags().GetString(flags.Format); format == "" {
				format = cli.Config.Format()
			}

			// Print out events in requested format
			if format == "json" {
				if err := helpers.PrintJSON(r, cmd.OutOrStdout()); err != nil {
					return err
				}
			} else {
				printToTable(r, cmd.OutOrStdout())
			}

			return nil
		},
	}

	helpers.AddFormatFlag(cmd.Flags())
	cmd.Flags().Bool(flags.AllOrgs, false, "Include records from all organizations")

	return cmd
}

func printToTable(queryResults []types.EventFilter, io io.Writer) {
	rows := make([]*table.Row, len(queryResults))
	for i, result := range queryResults {
		rows[i] = &table.Row{Value: result}
	}

	table := table.New([]*table.Column{
		{
			Title:       "Name",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				filter, _ := data.(types.EventFilter)
				return filter.Name
			},
		},
		{
			Title: "Action",
			CellTransformer: func(data interface{}) string {
				filter, _ := data.(types.EventFilter)
				return filter.Action
			},
		},
		{
			Title: "Statements",
			CellTransformer: func(data interface{}) string {
				filter, _ := data.(types.EventFilter)
				return strings.Join(filter.Statements, " && ")
			},
		},
	})

	table.Render(io, rows)
}

package check

import (
	"io"
	"strconv"
	"strings"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/table"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// ListCommand defines new list events command
func ListCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "list",
		Short:        "list checks",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			format, _ := cmd.Flags().GetString("format")

			// Fetch checks from the API
			r, err := cli.Client.ListChecks()
			if err != nil {
				return err
			}

			// Print out events in requested format
			if format == "json" {
				helpers.PrintJSON(r, cmd.OutOrStdout())
			} else {
				printCheckConfigsToTable(r, cmd.OutOrStdout())
			}

			return nil
		},
	}

	helpers.AddFormatFlag(cmd.Flags(), cli.Config)

	return cmd
}

func printCheckConfigsToTable(queryResults []types.CheckConfig, io io.Writer) {
	rows := make([]*table.Row, len(queryResults))
	for i, result := range queryResults {
		rows[i] = &table.Row{Value: result}
	}

	table := table.New([]*table.Column{
		{
			Title:       "Name",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				check, _ := data.(types.CheckConfig)
				return check.Name
			},
		},
		{
			Title: "Command",
			CellTransformer: func(data interface{}) string {
				check, _ := data.(types.CheckConfig)
				return check.Command
			},
		},
		{
			Title: "Interval",
			CellTransformer: func(data interface{}) string {
				check, _ := data.(types.CheckConfig)
				return strconv.Itoa(check.Interval)
			},
		},
		{
			Title: "Subscriptions",
			CellTransformer: func(data interface{}) string {
				check, _ := data.(types.CheckConfig)
				return strings.Join(check.Subscriptions, ",")
			},
		},
		{
			Title: "Handlers",
			CellTransformer: func(data interface{}) string {
				check, _ := data.(types.CheckConfig)
				return strings.Join(check.Handlers, ",")
			},
		},
		{
			Title: "Assets",
			CellTransformer: func(data interface{}) string {
				check, _ := data.(types.CheckConfig)
				return strings.Join(check.RuntimeAssets, ",")
			},
		},
	})

	table.Render(io, rows)
}

package check

import (
	"io"
	"strconv"
	"strings"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/flags"
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
			org := cli.Config.Organization()
			if ok, _ := cmd.Flags().GetBool(flags.AllOrgs); ok {
				org = "*"
			}

			// Fetch checks from the API
			r, err := cli.Client.ListChecks(org)
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
				printCheckConfigsToTable(r, cmd.OutOrStdout())
			}

			return nil
		},
	}

	helpers.AddFormatFlag(cmd.Flags(), cli.Config)
	cmd.Flags().Bool(flags.AllOrgs, false, "Include records from all organizations")

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
				interval := strconv.FormatUint(uint64(check.Interval), 10)
				return interval
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
		{
			Title: "Publish?",
			CellTransformer: func(data interface{}) string {
				check, _ := data.(types.CheckConfig)
				return strconv.FormatBool(check.Publish)
			},
		},
	})

	table.Render(io, rows)
}

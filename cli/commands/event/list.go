package event

import (
	"io"
	"strconv"
	"time"

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
		Short:        "list events",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			org := cli.Config.Organization()
			if ok, _ := cmd.Flags().GetBool(flags.AllOrgs); ok {
				org = "*"
			}

			// Fetch events from API
			r, err := cli.Client.ListEvents(org)
			if err != nil {
				return err
			}

			// Determine the format to use to output the data
			var format string
			if format = helpers.GetChangedStringValueFlag("format", cmd.Flags()); format == "" {
				format = cli.Config.Format()
			}

			if format == "json" {
				if err := helpers.PrintJSON(r, cmd.OutOrStdout()); err != nil {
					return err
				}
			} else {
				printEventsToTable(r, cmd.OutOrStdout())
			}

			return nil
		},
	}

	helpers.AddFormatFlag(cmd.Flags())
	cmd.Flags().Bool(flags.AllOrgs, false, "Include records from all organizations")

	return cmd
}

func printEventsToTable(queryResults []types.Event, io io.Writer) {
	rows := make([]*table.Row, len(queryResults))
	for i, result := range queryResults {
		rows[i] = &table.Row{Value: result}
	}

	table := table.New([]*table.Column{
		{
			Title:       "Entity",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				event, _ := data.(types.Event)
				return event.Entity.ID
			},
		},
		{
			Title: "Check",
			CellTransformer: func(data interface{}) string {
				event, _ := data.(types.Event)
				return event.Check.Config.Name
			},
		},
		{
			Title: "Output",
			CellTransformer: func(data interface{}) string {
				event, _ := data.(types.Event)
				return event.Check.Output
			},
		},
		{
			Title: "Status",
			CellTransformer: func(data interface{}) string {
				event, _ := data.(types.Event)
				return strconv.Itoa(int(event.Check.Status))
			},
		},
		{
			Title: "Timestamp",
			CellTransformer: func(data interface{}) string {
				event, _ := data.(types.Event)
				time := time.Unix(event.Timestamp, 0)
				return time.String()
			},
		},
	})

	table.Render(io, rows)
}

package event

import (
	"time"

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
		Short:        "list events",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			format, _ := cmd.Flags().GetString("format")

			// Fetch events from API
			r, err := cli.Client.ListEvents()
			if err != nil {
				return err
			}

			if format == "json" {
				helpers.PrintResultsToPrettyJSON(r)
			} else {
				printEventsToTable(r)
			}

			return nil
		},
	}

	helpers.AddFormatFlag(cmd.Flags(), cli.Config)

	return cmd
}

func printEventsToTable(queryResults []types.Event) {
	rows := make([]*table.Row, len(queryResults))
	for i, result := range queryResults {
		rows[i] = &table.Row{Value: result}
	}

	table := table.New([]*table.Column{
		{
			Title:       "Source",
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
				return event.Check.Name
			},
		},
		{
			Title: "Result",
			CellTransformer: func(data interface{}) string {
				event, _ := data.(types.Event)
				return event.Check.Output
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

	table.Render(rows)
}

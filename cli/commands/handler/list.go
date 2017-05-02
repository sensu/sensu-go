package handler

import (
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
		Short:        "list handlers",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			format, _ := cmd.Flags().GetString("format")

			// Fetch handlers from API
			r, err := cli.Client.ListHandlers()
			if err != nil {
				return err
			}

			if format == "json" {
				helpers.PrettyPrintResultsToJSON(r)
			} else {
				printHandlersToTable(r)
			}

			return nil
		},
	}

	helpers.AddFormatFlag(cmd.Flags(), cli)

	return cmd
}

func printHandlersToTable(queryResults []types.Handler) {
	rows := make([]*table.Row, len(queryResults))
	for i, result := range queryResults {
		rows[i] = &table.Row{Value: result}
	}

	table := table.New([]*table.Column{
		{
			Title:       "Name",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				handler, _ := data.(types.Handler)
				return handler.Name
			},
		},
		{
			Title: "Type",
			CellTransformer: func(data interface{}) string {
				handler, _ := data.(types.Handler)
				return handler.Type
			},
		},
		{
			Title: "Command",
			CellTransformer: func(data interface{}) string {
				handler, _ := data.(types.Handler)
				return handler.Command
			},
		},
		{
			Title: "Mutator",
			CellTransformer: func(data interface{}) string {
				handler, _ := data.(types.Handler)
				return handler.Mutator
			},
		},
		{
			Title: "Handlers",
			CellTransformer: func(data interface{}) string {
				handler, _ := data.(types.Handler)
				return strings.Join(handler.Handlers, ",")
			},
		},
	})

	table.Render(rows)
}

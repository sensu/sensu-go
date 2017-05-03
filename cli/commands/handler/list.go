package handler

import (
	"fmt"
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
				helpers.PrintResultsToPrettyJSON(r)
			} else {
				printHandlersToTable(r)
			}

			return nil
		},
	}

	helpers.AddFormatFlag(cmd.Flags(), cli.Config)

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
		{
			Title: "Socket",
			CellTransformer: func(data interface{}) string {
				handler, _ := data.(types.Handler)

				if len(handler.Socket.Host) > 0 {
					return fmt.Sprintf("%s://%s:%d", handler.Type, handler.Socket.Host, handler.Socket.Port)
				}

				return ""
			},
		},
	})

	table.Render(rows)
}

package entity

import (
	"io"
	"strings"
	"time"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/table"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// ListCommand defines new list entity command
func ListCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "list",
		Short:        "list entities",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			format, _ := cmd.Flags().GetString("format")

			// Fetch handlers from API
			r, err := cli.Client.ListEntities()
			if err != nil {
				return err
			}

			if format == "json" {
				helpers.PrintJSON(r, cmd.OutOrStdout())
			} else {
				printEntitiesToTable(r, cmd.OutOrStdout())
			}

			return nil
		},
	}

	helpers.AddFormatFlag(cmd.Flags(), cli.Config)

	return cmd
}

func printEntitiesToTable(queryResults []types.Entity, writer io.Writer) {
	rows := make([]*table.Row, len(queryResults))
	for i, result := range queryResults {
		rows[i] = &table.Row{Value: result}
	}

	table := table.New([]*table.Column{
		{
			Title:       "Host",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				entity, _ := data.(types.Entity)
				return entity.System.Hostname
			},
		},
		{
			Title: "OS",
			CellTransformer: func(data interface{}) string {
				entity, _ := data.(types.Entity)
				return entity.System.OS
			},
		},
		{
			Title: "Subscriptions",
			CellTransformer: func(data interface{}) string {
				entity, _ := data.(types.Entity)
				return strings.Join(entity.Subscriptions, ",")
			},
		},
		{
			Title: "Last Seen",
			CellTransformer: func(data interface{}) string {
				entity, _ := data.(types.Entity)
				time := time.Unix(entity.LastSeen*int64(time.Second), 0)
				return time.String()
			},
		},
	})

	table.Render(writer, rows)
}

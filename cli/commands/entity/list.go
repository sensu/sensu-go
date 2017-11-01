package entity

import (
	"io"
	"strings"
	"time"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/flags"
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
			org := cli.Config.Organization()
			if ok, _ := cmd.Flags().GetBool(flags.AllOrgs); ok {
				org = "*"
			}

			// Fetch handlers from API
			r, err := cli.Client.ListEntities(org)
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
				printEntitiesToTable(r, cmd.OutOrStdout())
			}

			return nil
		},
	}

	helpers.AddFormatFlag(cmd.Flags())
	cmd.Flags().Bool(flags.AllOrgs, false, "Include records from all organizations")

	return cmd
}

func printEntitiesToTable(queryResults []types.Entity, writer io.Writer) {
	rows := make([]*table.Row, len(queryResults))
	for i, result := range queryResults {
		rows[i] = &table.Row{Value: result}
	}

	table := table.New([]*table.Column{
		{
			Title:       "ID",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				entity, _ := data.(types.Entity)
				return entity.ID
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
				time := time.Unix(entity.LastSeen, 0)
				return time.String()
			},
		},
	})

	table.Render(writer, rows)
}

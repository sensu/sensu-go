package role

import (
	"io"

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
		Short:        "list roles",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Fetch roles from API
			r, err := cli.Client.ListRoles()
			if err != nil {
				return err
			}

			// Determine the format to use to output the data
			var format string
			if format, _ = cmd.Flags().GetString("format"); format == "" {
				format = cli.Config.Format()
			}

			if format == "json" {
				helpers.PrintJSON(r, cmd.OutOrStdout())
			} else {
				printRolesToTable(r, cmd.OutOrStdout())
			}

			return nil
		},
	}

	helpers.AddFormatFlag(cmd.Flags(), cli.Config)

	return cmd
}

func printRolesToTable(queryResults []types.Role, io io.Writer) {
	rows := make([]*table.Row, len(queryResults))
	for i, result := range queryResults {
		rows[i] = &table.Row{Value: result}
	}

	table := table.New([]*table.Column{
		{
			Title:       "Name",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				role, _ := data.(types.Role)
				return role.Name
			},
		},
	})

	table.Render(io, rows)
}

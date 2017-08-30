package organization

import (
	"io"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/table"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// ListCommand defines *organization list* command
func ListCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "list",
		Short:        "list organizations",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Fetch orgs from API
			r, err := cli.Client.ListOrganizations()
			if err != nil {
				return err
			}

			// Determine the format to use to output the data
			var format string
			if format, _ = cmd.Flags().GetString("format"); format == "" {
				format = cli.Config.Format()
			}

			if format == "json" {
				if err := helpers.PrintJSON(r, cmd.OutOrStdout()); err != nil {
					return err
				}
			} else {
				printOrganizationsToTable(r, cmd.OutOrStdout())
			}

			return nil
		},
	}

	helpers.AddFormatFlag(cmd.Flags(), cli.Config)

	return cmd
}

func printOrganizationsToTable(queryResults []types.Organization, io io.Writer) {
	rows := make([]*table.Row, len(queryResults))
	for i, result := range queryResults {
		rows[i] = &table.Row{Value: result}
	}

	table := table.New([]*table.Column{
		{
			Title:       "Name",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				org, _ := data.(types.Organization)
				return org.Name
			},
		},
		{
			Title: "Description",
			CellTransformer: func(data interface{}) string {
				org, _ := data.(types.Organization)
				return org.Description
			},
		},
	})

	table.Render(io, rows)
}

package organization

import (
	"errors"
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
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}
			// Fetch orgs from API
			results, err := cli.Client.ListOrganizations()
			if err != nil {
				return err
			}

			// Print the results based on the user preferences
			return helpers.Print(cmd, cli.Config.Format(), printToTable, results)
		},
	}

	helpers.AddFormatFlag(cmd.Flags())

	return cmd
}

func printToTable(results interface{}, writer io.Writer) {
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

	table.Render(writer, results)
}

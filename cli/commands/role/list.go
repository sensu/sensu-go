package role

import (
	"io"
	"strconv"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/table"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// ListCommand defines a command to list roles
func ListCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "list",
		Short:        "list Roles",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Fetch roles from API
			results, err := cli.Client.ListRoles()
			if err != nil {
				return err
			}
			// Print the results based on the user preferences
			resources := []types.Resource{}
			for i := range results {
				resources = append(resources, &results[i])
			}
			return helpers.Print(cmd, cli.Config.Format(), printToTable, resources, results)
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
				role, ok := data.(types.Role)
				if !ok {
					return cli.TypeError
				}
				return role.Name
			},
		},
		{
			Title: "Namespace",
			CellTransformer: func(data interface{}) string {
				role, ok := data.(types.Role)
				if !ok {
					return cli.TypeError
				}
				return role.Namespace
			},
		},
		{
			Title: "Rules",
			CellTransformer: func(data interface{}) string {
				role, ok := data.(types.Role)
				if !ok {
					return cli.TypeError
				}
				return strconv.Itoa(len(role.Rules))
			},
		},
	})
	table.Render(writer, results)
}

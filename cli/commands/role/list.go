package role

import (
	"io"
	"reflect"

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
			results, err := cli.Client.ListRoles()
			if err != nil {
				return err
			}

			// Print the results based on the user preferences
			helpers.Print(cmd, cli.Config.Format(), printRolesToTable, results)

			return nil
		},
	}

	helpers.AddFormatFlag(cmd.Flags())

	return cmd
}

func printRolesToTable(results interface{}, writer io.Writer) {
	if reflect.TypeOf(results).Kind() != reflect.Slice {
		return
	}
	slice := reflect.ValueOf(results)

	rows := make([]*table.Row, slice.Len())
	for i := 0; i < slice.Len(); i++ {
		rows[i] = &table.Row{Value: slice.Index(i).Interface()}
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

	table.Render(writer, rows)
}

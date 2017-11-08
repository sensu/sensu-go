package filter

import (
	"io"
	"reflect"
	"strings"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/flags"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/table"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// ListCommand defines the 'filter list' subcommand
func ListCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "list",
		Short:        "list filters",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			org := cli.Config.Organization()
			if ok, _ := cmd.Flags().GetBool(flags.AllOrgs); ok {
				org = "*"
			}

			// Fetch filters from the API
			results, err := cli.Client.ListFilters(org)
			if err != nil {
				return err
			}

			// Print the results based on the user preferences
			helpers.Print(cmd, cli.Config.Format(), printToTable, results)

			return nil
		},
	}

	helpers.AddFormatFlag(cmd.Flags())
	helpers.AddAllOrganization(cmd.Flags())

	return cmd
}

func printToTable(results interface{}, writer io.Writer) {
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
				filter, _ := data.(types.EventFilter)
				return filter.Name
			},
		},
		{
			Title: "Action",
			CellTransformer: func(data interface{}) string {
				filter, _ := data.(types.EventFilter)
				return filter.Action
			},
		},
		{
			Title: "Statements",
			CellTransformer: func(data interface{}) string {
				filter, _ := data.(types.EventFilter)
				return strings.Join(filter.Statements, " && ")
			},
		},
	})

	table.Render(writer, rows)
}

package filter

import (
	"errors"
	"io"
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
			if len(args) != 0 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}
			namespace := cli.Config.Namespace()
			if ok, _ := cmd.Flags().GetBool(flags.AllNamespaces); ok {
				namespace = types.NamespaceTypeAll
			}

			// Fetch filters from the API
			results, err := cli.Client.ListFilters(namespace)
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
	helpers.AddAllNamespace(cmd.Flags())

	return cmd
}

func printToTable(results interface{}, writer io.Writer) {
	table := table.New([]*table.Column{
		{
			Title:       "Name",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				filter, ok := data.(types.EventFilter)
				if !ok {
					return cli.TypeError
				}
				return filter.Name
			},
		},
		{
			Title: "Action",
			CellTransformer: func(data interface{}) string {
				filter, ok := data.(types.EventFilter)
				if !ok {
					return cli.TypeError
				}
				return filter.Action
			},
		},
		{
			Title: "Expressions",
			CellTransformer: func(data interface{}) string {
				filter, ok := data.(types.EventFilter)
				if !ok {
					return cli.TypeError
				}
				return strings.Join(filter.Expressions, " && ")
			},
		},
	})

	table.Render(writer, results)
}

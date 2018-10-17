package mutator

import (
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/flags"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/table"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// ListCommand defines the 'mutator list' subcommand
func ListCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "list",
		Short:        "list mutators",
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

			// Fetch mutators from the API
			results, err := cli.Client.ListMutators(namespace)
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
				mutator, ok := data.(types.Mutator)
				if !ok {
					return cli.TypeError
				}
				return mutator.Name
			},
		},
		{
			Title:       "Command",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				mutator, ok := data.(types.Mutator)
				if !ok {
					return cli.TypeError
				}
				return mutator.Command
			},
		},
		{
			Title:       "Environment Variables",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				mutator, ok := data.(types.Mutator)
				if !ok {
					return cli.TypeError
				}
				return strings.Join(mutator.EnvVars, ",")
			},
		},
		{
			Title:       "Timeout",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				mutator, ok := data.(types.Mutator)
				if !ok {
					return cli.TypeError
				}
				timeout := strconv.FormatUint(uint64(mutator.Timeout), 10)
				return timeout
			},
		},
	})
	table.Render(writer, results)
}

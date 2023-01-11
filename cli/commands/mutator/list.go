package mutator

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/cli/commands/flags"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/table"

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
				namespace = corev2.NamespaceTypeAll
			}

			opts, err := helpers.ListOptionsFromFlags(cmd.Flags())
			if err != nil {
				return err
			}

			// Fetch mutators from the API
			var header http.Header
			results := []corev2.Mutator{}
			err = cli.Client.List(client.MutatorsPath(namespace), &results, &opts, &header)
			if err != nil {
				return err
			}

			// Print the results based on the user preferences
			resources := []corev3.Resource{}
			for i := range results {
				resources = append(resources, &results[i])
			}
			return helpers.PrintList(cmd, cli.Config.Format(), printToTable, resources, results, header)
		},
	}

	helpers.AddFormatFlag(cmd.Flags())
	helpers.AddAllNamespace(cmd.Flags())
	helpers.AddFieldSelectorFlag(cmd.Flags())
	helpers.AddLabelSelectorFlag(cmd.Flags())
	helpers.AddChunkSizeFlag(cmd.Flags())

	return cmd
}

func printToTable(results interface{}, writer io.Writer) {
	table := table.New([]*table.Column{
		{
			Title:       "Name",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				mutator, ok := data.(corev2.Mutator)
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
				mutator, ok := data.(corev2.Mutator)
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
				mutator, ok := data.(corev2.Mutator)
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
				mutator, ok := data.(corev2.Mutator)
				if !ok {
					return cli.TypeError
				}
				timeout := strconv.FormatUint(uint64(mutator.Timeout), 10)
				return timeout
			},
		},
		{
			Title: "Assets",
			CellTransformer: func(data interface{}) string {
				mutator, ok := data.(corev2.Mutator)
				if !ok {
					return cli.TypeError
				}
				return strings.Join(mutator.RuntimeAssets, ",")
			},
		},
	})
	table.Render(writer, results)
}

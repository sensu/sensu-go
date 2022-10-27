package filter

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/cli/commands/flags"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/table"

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
				namespace = corev2.NamespaceTypeAll
			}

			opts, err := helpers.ListOptionsFromFlags(cmd.Flags())
			if err != nil {
				return err
			}

			// Fetch filters from the API
			var header http.Header
			results := []corev2.EventFilter{}
			err = cli.Client.List(client.FiltersPath(namespace), &results, &opts, &header)
			if err != nil {
				return err
			}

			// Print the results based on the user preferences
			resources := []corev2.Resource{}
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
				filter, ok := data.(corev2.EventFilter)
				if !ok {
					return cli.TypeError
				}
				return filter.Name
			},
		},
		{
			Title: "Action",
			CellTransformer: func(data interface{}) string {
				filter, ok := data.(corev2.EventFilter)
				if !ok {
					return cli.TypeError
				}
				return filter.Action
			},
		},
		{
			Title: "Expressions",
			CellTransformer: func(data interface{}) string {
				filter, ok := data.(corev2.EventFilter)
				if !ok {
					return cli.TypeError
				}
				var expressions []string
				var sep string
				switch action := filter.Action; action {
				case corev2.EventFilterActionAllow:
					sep = " && "
				case corev2.EventFilterActionDeny:
					sep = " || "
				default:
					sep = ", "
				}
				for _, exp := range filter.Expressions {
					expressions = append(expressions, fmt.Sprintf("(%s)", exp))
				}
				return strings.Join(expressions, sep)
			},
		},
	})

	table.Render(writer, results)
}

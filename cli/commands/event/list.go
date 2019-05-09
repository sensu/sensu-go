package event

import (
	"errors"
	"io"
	"strconv"
	"time"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/flags"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/globals"
	"github.com/sensu/sensu-go/cli/elements/table"
	"github.com/sensu/sensu-go/types"

	"github.com/spf13/cobra"
)

// ListCommand defines new list events command
func ListCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "list",
		Short:        "list events",
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

			opts, err := helpers.ListOptionsFromFlags(cmd.Flags())
			if err != nil {
				return err
			}

			// Fetch events from API
			results, err := cli.Client.ListEvents(namespace, &opts)
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
	helpers.AddFieldSelectorFlag(cmd.Flags())
	helpers.AddLabelSelectorFlag(cmd.Flags())

	return cmd
}

func printToTable(results interface{}, writer io.Writer) {
	table := table.New([]*table.Column{
		{
			Title:       "Entity",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				event, ok := data.(types.Event)
				if !ok {
					return cli.TypeError
				}
				return event.Entity.Name
			},
		},
		{
			Title: "Check",
			CellTransformer: func(data interface{}) string {
				event, ok := data.(types.Event)
				if !ok {
					return cli.TypeError
				}
				return event.Check.Name
			},
		},
		{
			Title: "Output",
			CellTransformer: func(data interface{}) string {
				event, ok := data.(types.Event)
				if !ok {
					return cli.TypeError
				}
				return event.Check.Output
			},
		},
		{
			Title: "Status",
			CellTransformer: func(data interface{}) string {
				event, ok := data.(types.Event)
				if !ok {
					return cli.TypeError
				}
				return strconv.Itoa(int(event.Check.Status))
			},
		},
		{
			Title: "Silenced",
			CellTransformer: func(data interface{}) string {
				event, ok := data.(types.Event)
				if !ok {
					return cli.TypeError
				}
				return globals.BooleanStyleP(len(event.Check.Silenced) > 0)
			},
		},
		{
			Title: "Timestamp",
			CellTransformer: func(data interface{}) string {
				event, ok := data.(types.Event)
				if !ok {
					return cli.TypeError
				}
				time := time.Unix(event.Timestamp, 0)
				return time.String()
			},
		},
	})

	table.Render(writer, results)
}

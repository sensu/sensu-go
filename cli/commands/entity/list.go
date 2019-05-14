package entity

import (
	"errors"
	"io"
	"strings"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/flags"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/commands/timeutil"
	"github.com/sensu/sensu-go/cli/elements/table"
	"github.com/sensu/sensu-go/types"

	"github.com/spf13/cobra"
)

// ListCommand defines new list entity command
func ListCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "list",
		Short:        "list entities",
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

			// Fetch handlers from API
			results, err := cli.Client.ListEntities(namespace, &opts)
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
	helpers.AddChunkSizeFlag(cmd.Flags())

	return cmd
}

func printToTable(results interface{}, writer io.Writer) {
	table := table.New([]*table.Column{
		{
			Title:       "ID",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				entity, ok := data.(types.Entity)
				if !ok {
					return cli.TypeError
				}
				return entity.Name
			},
		},
		{
			Title: "Class",
			CellTransformer: func(data interface{}) string {
				entity, ok := data.(types.Entity)
				if !ok {
					return cli.TypeError
				}
				return entity.EntityClass
			},
		},
		{
			Title: "OS",
			CellTransformer: func(data interface{}) string {
				entity, ok := data.(types.Entity)
				if !ok {
					return cli.TypeError
				}
				return entity.System.OS
			},
		},
		{
			Title: "Subscriptions",
			CellTransformer: func(data interface{}) string {
				entity, ok := data.(types.Entity)
				if !ok {
					return cli.TypeError
				}
				return strings.Join(entity.Subscriptions, ",")
			},
		},
		{
			Title: "Last Seen",
			CellTransformer: func(data interface{}) string {
				entity, ok := data.(types.Entity)
				if !ok {
					return cli.TypeError
				}
				return timeutil.HumanTimestamp(entity.LastSeen)
			},
		},
	})

	table.Render(writer, results)
}

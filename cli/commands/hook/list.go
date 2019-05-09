package hook

import (
	"errors"
	"io"
	"strconv"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/flags"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/table"
	"github.com/sensu/sensu-go/types"

	"github.com/spf13/cobra"
)

// ListCommand defines new list events command
func ListCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "list",
		Short:        "list hooks",
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

			// Fetch hooks from the API
			results, err := cli.Client.ListHooks(namespace, &opts)
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
			Title:       "Name",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				hook, ok := data.(types.HookConfig)
				if !ok {
					return cli.TypeError
				}
				return hook.Name
			},
		},
		{
			Title: "Command",
			CellTransformer: func(data interface{}) string {
				hook, ok := data.(types.HookConfig)
				if !ok {
					return cli.TypeError
				}
				return hook.Command
			},
		},
		{
			Title: "Timeout",
			CellTransformer: func(data interface{}) string {
				hook, ok := data.(types.HookConfig)
				if !ok {
					return cli.TypeError
				}
				timeout := strconv.FormatUint(uint64(hook.Timeout), 10)
				return timeout
			},
		},
		{
			Title: "Stdin?",
			CellTransformer: func(data interface{}) string {
				hook, ok := data.(types.HookConfig)
				if !ok {
					return cli.TypeError
				}
				return strconv.FormatBool(hook.Stdin)
			},
		},
	})

	table.Render(writer, results)
}

package hook

import (
	"errors"
	"io"
	"net/http"
	"strconv"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/cli/commands/flags"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/table"

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
				namespace = corev2.NamespaceTypeAll
			}

			opts, err := helpers.ListOptionsFromFlags(cmd.Flags())
			if err != nil {
				return err
			}

			// Fetch hooks from the API
			var header http.Header
			results := []corev2.HookConfig{}
			err = cli.Client.List(client.HooksPath(namespace), &results, &opts, &header)
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
				hook, ok := data.(corev2.HookConfig)
				if !ok {
					return cli.TypeError
				}
				return hook.Name
			},
		},
		{
			Title: "Command",
			CellTransformer: func(data interface{}) string {
				hook, ok := data.(corev2.HookConfig)
				if !ok {
					return cli.TypeError
				}
				return hook.Command
			},
		},
		{
			Title: "Timeout",
			CellTransformer: func(data interface{}) string {
				hook, ok := data.(corev2.HookConfig)
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
				hook, ok := data.(corev2.HookConfig)
				if !ok {
					return cli.TypeError
				}
				return strconv.FormatBool(hook.Stdin)
			},
		},
	})

	table.Render(writer, results)
}

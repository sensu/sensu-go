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
			org := cli.Config.Organization()
			if ok, _ := cmd.Flags().GetBool(flags.AllOrgs); ok {
				org = "*"
			}

			// Fetch hooks from the API
			results, err := cli.Client.ListHooks(org)
			if err != nil {
				return err
			}

			// Print the results based on the user preferences
			return helpers.Print(cmd, cli.Config.Format(), printToTable, results)
		},
	}

	helpers.AddFormatFlag(cmd.Flags())
	helpers.AddAllOrganization(cmd.Flags())

	return cmd
}

func printToTable(results interface{}, writer io.Writer) {
	table := table.New([]*table.Column{
		{
			Title:       "Name",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				hook, _ := data.(types.HookConfig)
				return hook.Name
			},
		},
		{
			Title: "Command",
			CellTransformer: func(data interface{}) string {
				hook, _ := data.(types.HookConfig)
				return hook.Command
			},
		},
		{
			Title: "Timeout",
			CellTransformer: func(data interface{}) string {
				hook, _ := data.(types.HookConfig)
				timeout := strconv.FormatUint(uint64(hook.Timeout), 10)
				return timeout
			},
		},
		{
			Title: "Stdin?",
			CellTransformer: func(data interface{}) string {
				hook, _ := data.(types.HookConfig)
				return strconv.FormatBool(hook.Stdin)
			},
		},
	})

	table.Render(writer, results)
}

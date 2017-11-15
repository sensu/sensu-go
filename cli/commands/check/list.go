package check

import (
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

// ListCommand defines new list events command
func ListCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "list",
		Short:        "list checks",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			org := cli.Config.Organization()
			if ok, _ := cmd.Flags().GetBool(flags.AllOrgs); ok {
				org = "*"
			}

			// Fetch checks from the API
			results, err := cli.Client.ListChecks(org)
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
	table := table.New([]*table.Column{
		{
			Title:       "Name",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				check, _ := data.(types.CheckConfig)
				return check.Name
			},
		},
		{
			Title: "Command",
			CellTransformer: func(data interface{}) string {
				check, _ := data.(types.CheckConfig)
				return check.Command
			},
		},
		{
			Title: "Interval",
			CellTransformer: func(data interface{}) string {
				check, _ := data.(types.CheckConfig)
				interval := strconv.FormatUint(uint64(check.Interval), 10)
				return interval
			},
		},
		{
			Title: "Subscriptions",
			CellTransformer: func(data interface{}) string {
				check, _ := data.(types.CheckConfig)
				return strings.Join(check.Subscriptions, ",")
			},
		},
		{
			Title: "Handlers",
			CellTransformer: func(data interface{}) string {
				check, _ := data.(types.CheckConfig)
				return strings.Join(check.Handlers, ",")
			},
		},
		{
			Title: "Assets",
			CellTransformer: func(data interface{}) string {
				check, _ := data.(types.CheckConfig)
				return strings.Join(check.RuntimeAssets, ",")
			},
		},
		{
			Title: "Publish?",
			CellTransformer: func(data interface{}) string {
				check, _ := data.(types.CheckConfig)
				return strconv.FormatBool(check.Publish)
			},
		},
	})

	table.Render(writer, results)
}

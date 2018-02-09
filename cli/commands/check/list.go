package check

import (
	"errors"
	"io"
	"strconv"
	"strings"

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
		Short:        "list checks",
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

			// Fetch checks from the API
			results, err := cli.Client.ListChecks(org)
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
			Title: "Cron",
			CellTransformer: func(data interface{}) string {
				check, _ := data.(types.CheckConfig)
				return check.Cron
			},
		},
		{
			Title: "Timeout",
			CellTransformer: func(data interface{}) string {
				check, _ := data.(types.CheckConfig)
				timeout := strconv.FormatUint(uint64(check.Timeout), 10)
				return timeout
			},
		},
		{
			Title: "TTL",
			CellTransformer: func(data interface{}) string {
				check, _ := data.(types.CheckConfig)
				ttl := strconv.FormatUint(uint64(check.Ttl), 10)
				return ttl
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
			Title: "Hooks",
			CellTransformer: func(data interface{}) string {
				check, _ := data.(types.CheckConfig)
				return globals.FormatHookLists(check.CheckHooks)
			},
		},
		{
			Title: "Publish?",
			CellTransformer: func(data interface{}) string {
				check, _ := data.(types.CheckConfig)
				return globals.BooleanStyleP(check.Publish)
			},
		},
		{
			Title: "Stdin?",
			CellTransformer: func(data interface{}) string {
				check, _ := data.(types.CheckConfig)
				return strconv.FormatBool(check.Stdin)
			},
		},
	})

	table.Render(writer, results)
}

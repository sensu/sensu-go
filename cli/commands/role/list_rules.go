package role

import (
	"errors"
	"io"
	"strings"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/table"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// ListRulesCommand defines new command to list rules associated w/ a role
func ListRulesCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "list-rules [ROLE]",
		Short:        "list rules associated with a role",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			// Fetch roles from API
			r, err := cli.Client.FetchRole(args[0])
			if err != nil {
				return err
			}

			// Determine the format to use to output the data
			var format string
			if format = helpers.GetChangedStringValueFlag("format", cmd.Flags()); format == "" {
				format = cli.Config.Format()
			}

			if format == "json" {
				if err := helpers.PrintJSON(r, cmd.OutOrStdout()); err != nil {
					return err
				}
			} else {
				printRulesToTable(r, cmd.OutOrStdout())
			}

			return nil
		},
	}

	helpers.AddFormatFlag(cmd.Flags())

	return cmd
}

func printRulesToTable(queryResults *types.Role, io io.Writer) {
	table := table.New([]*table.Column{
		{
			Title:       "Type",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				rule, _ := data.(types.Rule)
				return rule.Type
			},
		},
		{
			Title: "Org.",
			CellTransformer: func(data interface{}) string {
				rule, _ := data.(types.Rule)
				return rule.Organization
			},
		},
		{
			Title: "Env.",
			CellTransformer: func(data interface{}) string {
				rule, _ := data.(types.Rule)
				return rule.Environment
			},
		},
		{
			Title: "Permissions",
			CellTransformer: func(data interface{}) string {
				rule, _ := data.(types.Rule)
				return strings.Join(rule.Permissions, ",")
			},
		},
	})

	table.Render(io, queryResults.Rules)
}

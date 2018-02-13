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
			org := cli.Config.Organization()
			if ok, _ := cmd.Flags().GetBool(flags.AllOrgs); ok {
				org = "*"
			}

			// Fetch handlers from API
			results, err := cli.Client.ListEntities(org)
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
			Title:       "ID",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				entity, _ := data.(types.Entity)
				return entity.ID
			},
		},
		{
			Title: "Class",
			CellTransformer: func(data interface{}) string {
				entity, _ := data.(types.Entity)
				return entity.Class
			},
		},
		{
			Title: "OS",
			CellTransformer: func(data interface{}) string {
				entity, _ := data.(types.Entity)
				return entity.System.OS
			},
		},
		{
			Title: "Subscriptions",
			CellTransformer: func(data interface{}) string {
				entity, _ := data.(types.Entity)
				return strings.Join(entity.Subscriptions, ",")
			},
		},
		{
			Title: "Last Seen",
			CellTransformer: func(data interface{}) string {
				entity, _ := data.(types.Entity)
				return timeutil.HumanTimestamp(entity.LastSeen)
			},
		},
	})

	table.Render(writer, results)
}

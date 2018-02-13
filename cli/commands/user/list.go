package user

import (
	"errors"
	"io"
	"strings"

	"github.com/sensu/sensu-go/cli"
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
		Short:        "list users",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}
			// Fetch users from API
			results, err := cli.Client.ListUsers()
			if err != nil {
				return err
			}

			// Print the results based on the user preferences
			return helpers.Print(cmd, cli.Config.Format(), printToTable, results)
		},
	}

	helpers.AddFormatFlag(cmd.Flags())

	return cmd
}

func printToTable(results interface{}, writer io.Writer) {
	table := table.New([]*table.Column{
		{
			Title:       "Username",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				user, _ := data.(types.User)
				return user.Username
			},
		},
		{
			Title: "Roles",
			CellTransformer: func(data interface{}) string {
				user, _ := data.(types.User)
				return strings.Join(user.Roles, ",")
			},
		},
		{
			Title: "Enabled",
			CellTransformer: func(data interface{}) string {
				user, _ := data.(types.User)
				return globals.BooleanStyleP(!user.Disabled)
			},
		},
	})

	table.Render(writer, results)
}

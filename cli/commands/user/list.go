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

			opts, err := helpers.ListOptionsFromFlags(cmd.Flags())
			if err != nil {
				return err
			}

			// Fetch users from API
			results, header, err := cli.Client.ListUsers(opts)
			if err != nil {
				return err
			}

			err = helpers.PrintTitle(helpers.GetChangedStringValueFlag("format", cmd.Flags()), cli.Config.Format(), header, cmd.OutOrStdout())
			if err != nil {
				return err
			}

			resources := []types.Resource{}
			for i := range results {
				resources = append(resources, &results[i])
			}

			// Print the results based on the user preferences
			return helpers.Print(cmd, cli.Config.Format(), printToTable, resources, results)
		},
	}

	helpers.AddFormatFlag(cmd.Flags())
	helpers.AddFieldSelectorFlag(cmd.Flags())
	helpers.AddLabelSelectorFlag(cmd.Flags())

	return cmd
}

func printToTable(results interface{}, writer io.Writer) {
	table := table.New([]*table.Column{
		{
			Title:       "Username",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				user, ok := data.(types.User)
				if !ok {
					return cli.TypeError
				}
				return user.Username
			},
		},
		{
			Title: "Groups",
			CellTransformer: func(data interface{}) string {
				user, ok := data.(types.User)
				if !ok {
					return cli.TypeError
				}
				return strings.Join(user.Groups, ",")
			},
		},
		{
			Title: "Enabled",
			CellTransformer: func(data interface{}) string {
				user, ok := data.(types.User)
				if !ok {
					return cli.TypeError
				}
				return globals.BooleanStyleP(!user.Disabled)
			},
		},
	})

	table.Render(writer, results)
}

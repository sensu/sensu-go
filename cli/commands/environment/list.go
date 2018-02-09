package environment

import (
	"errors"
	"io"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/flags"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/table"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// ListCommand defines the environment list command
func ListCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "list",
		Short:        "list environments",
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

			// Fetch orgs from API
			results, err := cli.Client.ListEnvironments(org)
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
			Title:       "Organization",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				env, _ := data.(types.Environment)
				return env.Organization
			},
		},
		{
			Title:       "Name",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				env, _ := data.(types.Environment)
				return env.Name
			},
		},
		{
			Title: "Description",
			CellTransformer: func(data interface{}) string {
				env, _ := data.(types.Environment)
				return env.Description
			},
		},
	})

	table.Render(writer, results)
}

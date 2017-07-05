package organization

import (
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// DeleteCommand adds a command that allows user to delete users
func DeleteCommand(cli *cli.SensuCli) *cobra.Command {
	return &cobra.Command{
		Use:          "delete [ORGANIZATION]",
		Short:        "delete provided organization",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no name is present print out usage
			if len(args) != 1 {
				cmd.Help()
				return nil
			}

			org := args[0]
			err := cli.Client.DeleteOrganization(org)
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Deleted")
			return nil
		},
	}
}

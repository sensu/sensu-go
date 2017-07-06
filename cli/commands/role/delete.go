package role

import (
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// DeleteCommand defines new command to delete roles
func DeleteCommand(cli *cli.SensuCli) *cobra.Command {
	return &cobra.Command{
		Use:          "delete NAME",
		Short:        "delete role given name",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no name is present print out usage
			if len(args) != 1 {
				cmd.Help()
				return nil
			}

			name := args[0]
			err := cli.Client.DeleteRole(name)
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Deleted")
			return nil
		},
	}
}

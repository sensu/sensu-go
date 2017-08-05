package environment

import (
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// DeleteCommand adds a command that allows user to delete environments
func DeleteCommand(cli *cli.SensuCli) *cobra.Command {
	return &cobra.Command{
		Use:          "delete [ENVIRONMENT]",
		Short:        "delete specified environment",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no name is present print out usage
			if len(args) != 1 || args[0] == "" {
				cmd.Help()
				return nil
			}

			org := cli.Config.Organization()
			env := args[0]
			err := cli.Client.DeleteEnvironment(org, env)
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Deleted")
			return nil
		},
	}
}

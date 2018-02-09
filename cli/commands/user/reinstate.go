package user

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// ReinstateCommand adds a command that allows user to delete users
func ReinstateCommand(cli *cli.SensuCli) *cobra.Command {
	return &cobra.Command{
		Use:          "reinstate [USERNAME]",
		Short:        "reinstate disabled user given username",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no name is present print out usage
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			username := args[0]
			err := cli.Client.ReinstateUser(username)
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Reinstated")
			return nil
		},
	}
}

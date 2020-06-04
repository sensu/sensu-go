package user

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/backend/authentication/bcrypt"
	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// HashPasswordCommand adds a command that allows user to delete users
func HashPasswordCommand(cli *cli.SensuCli) *cobra.Command {
	return &cobra.Command{
		Use:          "hash-password [PASSWORD]",
		Short:        "generate a hash with the given password to use in the password_hash field",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no password is present print out usage
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("a password is required")
			}

			hash, err := bcrypt.HashPassword(args[0])
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), hash)
			return nil
		},
	}
}

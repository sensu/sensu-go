package user

import (
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/hooks"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// CreateCommand adds command that allows user to create new users
func CreateCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create NAME",
		Short:        "create new users",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			flags := cmd.Flags()
			isInteractive := flags.NFlag() == 0
			opts := newUserOpts()

			if isInteractive {
				opts.administerQuestionnaire()
			} else {
				opts.withFlags(flags)
				if len(args) > 0 {
					opts.Username = args[0]
				}
			}

			user := types.User{}
			opts.Copy(&user)

			if err := user.Validate(); err != nil {
				if !isInteractive {
					cmd.SilenceUsage = false
				}
				return err
			}

			err := cli.Client.CreateUser(&user)
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Created")
			return nil
		},
		Annotations: map[string]string{
			// We want to be able to run this command regardless of whether the CLI
			// has been configured.
			hooks.ConfigurationRequirement: hooks.ConfigurationNotRequired,
		},
	}

	cmd.Flags().StringP("username", "u", "", "Username")
	cmd.Flags().StringP("password", "p", "", "Password")
	cmd.Flags().Bool("admin", false, "Give user the administrator role")
	cmd.Flags().StringP("roles", "r", "", "Comma separated list of roles to assign")

	// Mark flags are required for bash-completions
	cmd.MarkFlagRequired("username")
	cmd.MarkFlagRequired("password")

	return cmd
}

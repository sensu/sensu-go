package user

import (
	"errors"
	"fmt"

	"github.com/AlecAivazis/survey"
	"github.com/sensu/sensu-go/backend/authentication/bcrypt"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/flags"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type resetPasswordOpts struct {
	New     string `survey:"new-password"`
	Confirm string `survey:"confirm-password"`
}

// ResetPasswordCommand adds command that allows user to reset the password of a
// user
func ResetPasswordCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "reset-password [USERNAME]",
		Short:        "reset password for given user",
		SilenceUsage: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			isInteractive, _ := cmd.Flags().GetBool(flags.Interactive)
			if !isInteractive {
				// Mark flags are required for bash-completions
				_ = cmd.MarkFlagRequired("new-password")
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("a username is required")
			}

			isInteractive, _ := cmd.Flags().GetBool(flags.Interactive)

			username := args[0]
			password := &resetPasswordOpts{}

			if isInteractive {
				// Prompt user for new password
				if err := password.administerQuestionnaire(); err != nil {
					return err
				}
			} else {
				if err := password.withFlags(cmd.Flags()); err != nil {
					_ = cmd.Help()
					return err
				}
			}

			// Validate new password
			if err := password.validate(); err != nil {
				return err
			}

			passwordHash, err := bcrypt.HashPassword(password.New)
			if err != nil {
				return err
			}

			// Reset password
			if err := cli.Client.ResetPassword(username, passwordHash); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Updated")
			return nil
		},
	}

	_ = cmd.Flags().StringP("password", "p", "", "new password")

	helpers.AddInteractiveFlag(cmd.Flags())

	return cmd
}

func (opts *resetPasswordOpts) administerQuestionnaire() error {
	qs := []*survey.Question{
		{
			Name:     "new-password",
			Prompt:   &survey.Password{Message: "New Password:\t\t"},
			Validate: survey.Required,
		},
		{
			Name:     "confirm-password",
			Prompt:   &survey.Password{Message: "Confirm Password:\t"},
			Validate: survey.Required,
		},
	}

	return survey.Ask(qs, opts)
}

func (opts *resetPasswordOpts) withFlags(flags *pflag.FlagSet) error {
	password, _ := flags.GetString("password")
	if password == "" {
		return errors.New("new password must be provided")
	}

	opts.New = password
	opts.Confirm = password

	return nil
}

func (opts *resetPasswordOpts) validate() error {
	if opts.New != opts.Confirm {
		return errPasswordsDoNotMatch
	}

	user := types.User{Password: opts.New}
	return user.ValidatePassword()
}

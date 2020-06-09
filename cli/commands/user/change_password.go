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

var (
	errEmptyCurrentPassword = errors.New("current user's password must be provided")
	errPasswordsDoNotMatch  = errors.New("given passwords do not match")
)

type passwordOpts struct {
	Current string `survey:"current-password"`
	New     string `survey:"new-password"`
	Confirm string `survey:"confirm-password"`
}

// SetPasswordCommand adds command that allows user to create new users
func SetPasswordCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "change-password [USERNAME]",
		Short:        "change password for given user",
		SilenceUsage: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			isInteractive, _ := cmd.Flags().GetBool(flags.Interactive)
			if !isInteractive {
				// Mark flags are required for bash-completions
				_ = cmd.MarkFlagRequired("new-password")
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			isInteractive, _ := cmd.Flags().GetBool(flags.Interactive)

			password := &passwordOpts{}
			var username string

			// Retrieve current username from JWT
			currentUsername := helpers.GetCurrentUsername(cli.Config)

			// If no username is given we use the current user's name
			if len(args) > 0 {
				username = args[0]
			} else {
				username = currentUsername
			}

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

			hash, err := bcrypt.HashPassword(password.New)
			if err != nil {
				return err
			}

			// Update password
			if err := cli.Client.UpdatePassword(username, hash, password.Current); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Updated")
			return nil
		},
	}

	_ = cmd.Flags().StringP("current-password", "c", "", "current password")
	_ = cmd.Flags().StringP("new-password", "p", "", "new password")

	helpers.AddInteractiveFlag(cmd.Flags())

	return cmd
}

func (opts *passwordOpts) administerQuestionnaire() error {
	qs := []*survey.Question{
		{
			Name:     "current-password",
			Prompt:   &survey.Password{Message: "Current Password:\t"},
			Validate: survey.Required,
		},
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

func (opts *passwordOpts) withFlags(flags *pflag.FlagSet) error {
	currentPassword, _ := flags.GetString("current-password")
	if currentPassword == "" {
		return errors.New("current password must be provided")
	}

	password, _ := flags.GetString("new-password")
	if password == "" {
		return errors.New("new password must be provided")
	}

	opts.Current = currentPassword
	opts.New = password
	opts.Confirm = password

	return nil
}

func (opts *passwordOpts) validate() error {
	if opts.Current == "" {
		return errEmptyCurrentPassword
	}

	if opts.New != opts.Confirm {
		return errPasswordsDoNotMatch
	}

	user := types.User{Password: opts.New}
	return user.ValidatePassword()
}

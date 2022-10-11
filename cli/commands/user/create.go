package user

import (
	"errors"
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/flags"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	corev2 "github.com/sensu/core/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type createOpts struct {
	Username             string `survey:"username"`
	Password             string `survey:"password"`
	PasswordConfirmation string `survey:"passwordConfirmation"`
	Groups               string `survey:"group"`
}

// CreateCommand adds command that allows user to create new users
func CreateCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create [NAME]",
		Short:        "create new users",
		SilenceUsage: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			isInteractive, _ := cmd.Flags().GetBool(flags.Interactive)
			if !isInteractive {
				// Mark flags are required for bash-completions
				_ = cmd.MarkFlagRequired("password")
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			isInteractive, _ := cmd.Flags().GetBool(flags.Interactive)
			opts := &createOpts{}

			if len(args) > 0 {
				opts.Username = args[0]
			}

			if isInteractive {
				if err := opts.administerQuestionnaire(); err != nil {
					return err
				}
			} else {
				opts.withFlags(cmd.Flags())
			}

			if isInteractive && opts.Password != opts.PasswordConfirmation {
				//lint:ignore ST1005 this error is written to stdout/stderr
				return errors.New("Password confirmation doesn't match the password")
			}
			user := opts.toUser()
			if err := user.Validate(); err != nil {
				if !isInteractive {
					cmd.SilenceUsage = false
				}
				return err
			}

			err := cli.Client.CreateUser(user)
			if err != nil {
				return err
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Created")
			return err
		},
	}

	_ = cmd.Flags().StringP("password", "p", "", "Password")
	_ = cmd.Flags().StringP("groups", "g", "", "Comma separated list of the groups to assign")

	helpers.AddInteractiveFlag(cmd.Flags())
	return cmd
}

func (opts *createOpts) withFlags(flags *pflag.FlagSet) {
	opts.Password, _ = flags.GetString("password")
	opts.Groups, _ = flags.GetString("groups")
}

func (opts *createOpts) administerQuestionnaire() error {
	var qs = []*survey.Question{
		{
			Name: "username",
			Prompt: &survey.Input{
				Message: "Username:",
				Default: opts.Username,
			},
			Validate: survey.Required,
		},
		{
			Name: "password",
			Prompt: &survey.Password{
				Message: "Password:",
			},
			Validate: survey.Required,
		},
		{
			Name: "passwordConfirmation",
			Prompt: &survey.Password{
				Message: "Retype password:",
			},
			Validate: survey.Required,
		},
		{
			Name: "groups",
			Prompt: &survey.Input{
				Message: "Groups:",
			},
		},
	}

	return survey.Ask(qs, opts)
}

func (opts *createOpts) toUser() *corev2.User {
	groups := helpers.SafeSplitCSV(opts.Groups)

	return &corev2.User{
		Username: opts.Username,
		Password: opts.Password,
		Groups:   groups,
	}
}

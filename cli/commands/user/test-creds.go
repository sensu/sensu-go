package user

import (
	"errors"

	"github.com/AlecAivazis/survey/v2"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/flags"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type testCredsOpts struct {
	Username string `survey:"username"`
	Password string `survey:"password"`
}

// TestCredsCommand adds a command that allows user to test other
// user's credentials.
func TestCredsCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "test-creds [NAME]",
		Short:        "test user credentials",
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
			opts := &testCredsOpts{}

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

			err := cli.Client.TestCreds(opts.Username, opts.Password)
			if err != nil {
				return err
			}

			return nil
		},
	}

	_ = cmd.Flags().StringP("password", "p", "", "Password")

	helpers.AddInteractiveFlag(cmd.Flags())
	return cmd
}

func (opts *testCredsOpts) withFlags(flags *pflag.FlagSet) {
	opts.Password, _ = flags.GetString("password")
}

func (opts *testCredsOpts) administerQuestionnaire() error {
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
	}

	return survey.Ask(qs, opts)
}

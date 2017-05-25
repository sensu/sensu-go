package user

import (
	"fmt"

	"github.com/AlecAivazis/survey"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type createOpts struct {
	Username string `survey:"username"`
	Password string `survey:"password"`
}

// CreateCommand adds command that allows user to create new users
func CreateCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create [NAME]",
		Short:        "create new users",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			flags := cmd.Flags()
			isInteractive := flags.NFlag() == 0
			opts := &createOpts{}

			if isInteractive {
				opts.administerQuestionnaire()
			} else {
				opts.withFlags(flags)
				if len(args) > 0 {
					opts.Username = args[0]
				}
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

			fmt.Fprintf(cmd.OutOrStdout(), "Created")
			return nil
		},
	}

	cmd.Flags().StringP("username", "u", "", "Username")
	cmd.Flags().StringP("password", "p", "", "Password")

	return cmd
}

func (opts *createOpts) withFlags(flags *pflag.FlagSet) {
	opts.Username, _ = flags.GetString("username")
	opts.Password, _ = flags.GetString("password")
}

func (opts *createOpts) administerQuestionnaire() {
	var qs = []*survey.Question{
		{
			Name: "username",
			Prompt: &survey.Input{
				Message: "Username:",
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

	survey.Ask(qs, opts)
}

func (opts *createOpts) toUser() *types.User {
	return &types.User{
		Username: opts.Username,
		Password: opts.Password,
	}
}

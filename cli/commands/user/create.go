package user

import (
	"errors"
	"fmt"

	"github.com/AlecAivazis/survey"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/flags"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type createOpts struct {
	Username string `survey:"username"`
	Password string `survey:"password"`
	Roles    string `survey:"roles"`
	Admin    bool
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
	_ = cmd.Flags().Bool("admin", false, "Give user the administrator role")
	_ = cmd.Flags().StringP("roles", "r", "", "Comma separated list of roles to assign")

	helpers.AddInteractiveFlag(cmd.Flags())
	return cmd
}

func (opts *createOpts) withFlags(flags *pflag.FlagSet) {
	opts.Password, _ = flags.GetString("password")
	opts.Roles, _ = flags.GetString("roles")

	if isAdmin, _ := flags.GetBool("admin"); isAdmin {
		opts.Admin = isAdmin
	}
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
			Name: "roles",
			Prompt: &survey.Input{
				Message: "Roles:",
			},
		},
	}

	return survey.Ask(qs, opts)
}

func (opts *createOpts) toUser() *types.User {
	roles := helpers.SafeSplitCSV(opts.Roles)

	if opts.Admin {
		roles = append(roles, "admin")
	}

	return &types.User{
		Username: opts.Username,
		Password: opts.Password,
		Roles:    roles,
	}
}

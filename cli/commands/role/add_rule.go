package role

import (
	"errors"
	"fmt"

	"github.com/AlecAivazis/survey"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type ruleOpts struct {
	Role        string   `survey:"role"`
	Type        string   `survey:"type"`
	Permissions []string `survey:"permissions"`
	Env         string
	Org         string
}

// AddRuleCommand defines new command to add rules to a role
func AddRuleCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "add-rule ROLE-NAME",
		Short:        "add-rule to role",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			flags := cmd.Flags()
			isInteractive := flags.NFlag() == 0

			opts := &ruleOpts{}

			if isInteractive {
				cmd.SilenceUsage = false
				if err := opts.administerQuestionnaire(); err != nil {
					return err
				}
			} else {
				opts.Role = args[0]
				opts.withFlags(flags)
			}

			if opts.Org == "" {
				opts.Org = cli.Config.Organization()
			}

			if opts.Env == "" {
				opts.Env = cli.Config.Environment()
			}

			if opts.Role == "" {
				return errors.New("must provide name of associated role")
			}

			// Instantiate rule from input
			rule := types.Rule{}
			opts.Copy(&rule)

			// Ensure that the given rule is valid
			if err := rule.Validate(); err != nil {
				return err
			}

			if err := cli.Client.AddRule(opts.Role, &rule); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Added")
			return nil
		},
	}

	cmd.Flags().StringP("type", "t", "", "type associated with the rule")
	cmd.Flags().BoolP("create", "c", false, "create permission")
	cmd.Flags().BoolP("read", "r", false, "read permission")
	cmd.Flags().BoolP("update", "u", false, "update permission")
	cmd.Flags().BoolP("delete", "d", false, "delete permission")

	// Mark flags are required for bash-completions
	cmd.MarkFlagRequired("type")

	return cmd
}

func (opts *ruleOpts) withFlags(flags *pflag.FlagSet) {
	opts.Type, _ = flags.GetString("type")
	opts.Org, _ = flags.GetString("organization")
	opts.Env, _ = flags.GetString("environment")

	if create, _ := flags.GetBool("create"); create {
		opts.Permissions = append(opts.Permissions, "create")
	}
	if read, _ := flags.GetBool("read"); read {
		opts.Permissions = append(opts.Permissions, "read")
	}
	if update, _ := flags.GetBool("update"); update {
		opts.Permissions = append(opts.Permissions, "update")
	}
	if delete, _ := flags.GetBool("delete"); delete {
		opts.Permissions = append(opts.Permissions, "delete")
	}
}

func (opts *ruleOpts) administerQuestionnaire() error {
	var qs = []*survey.Question{
		{
			Name: "role",
			Prompt: &survey.Input{
				Message: "Role Name:",
				Default: "",
			},
			Validate: survey.Required,
		},
		{
			Name: "type",
			Prompt: &survey.Input{
				Message: "Rule Type:",
				Default: "",
			},
			Validate: survey.Required,
		},
		{
			Name: "permissions",
			Prompt: &survey.MultiSelect{
				Message: "Permissions:",
				Options: []string{"create", "read", "update", "delete"},
			},
		},
	}

	return survey.Ask(qs, opts)
}

func (opts *ruleOpts) Copy(rule *types.Rule) {
	rule.Type = opts.Type
	rule.Environment = opts.Env
	rule.Organization = opts.Org
	rule.Permissions = opts.Permissions
}

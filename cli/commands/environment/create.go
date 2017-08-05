package environment

import (
	"fmt"

	"github.com/AlecAivazis/survey"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/hooks"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type createOpts struct {
	Description  string `survey:"description"`
	Name         string `survey:"name"`
	Organization string `survey:"organization"`
}

// CreateCommand adds command that allows users to create new environments
func CreateCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create NAME",
		Short:        "create new environment",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			flags := cmd.Flags()
			isInteractive := flags.NFlag() == 0
			opts := &createOpts{}
			opts.Organization = cli.Config.Organization()

			if isInteractive {
				opts.administerQuestionnaire()
			} else {
				opts.withFlags(flags)
				if len(args) > 0 {
					opts.Name = args[0]
				}
			}

			org, env := opts.toEnvironment()
			if org == "" {
				return fmt.Errorf("an organization must be provided")
			}

			if err := env.Validate(); err != nil {
				if !isInteractive {
					cmd.SilenceUsage = false
				}
				return err
			}

			if err := cli.Client.CreateEnvironment(org, env); err != nil {
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

	cmd.Flags().StringP("description", "", "", "Description of environment")
	cmd.Flags().StringP("name", "", "", "Name of environment")

	// Mark flags are required for bash-completions
	cmd.MarkFlagRequired("name")

	return cmd
}

func (opts *createOpts) withFlags(flags *pflag.FlagSet) {
	opts.Description, _ = flags.GetString("description")
	opts.Name, _ = flags.GetString("name")
}

func (opts *createOpts) administerQuestionnaire() {
	var qs = []*survey.Question{
		{
			Name: "name",
			Prompt: &survey.Input{
				Message: "Name:",
			},
			Validate: survey.Required,
		},
		{
			Name: "description",
			Prompt: &survey.Input{
				Message: "Description:",
			},
		},
	}

	survey.Ask(qs, opts)
}

func (opts *createOpts) toEnvironment() (string, *types.Environment) {
	return opts.Organization, &types.Environment{
		Description: opts.Description,
		Name:        opts.Name,
	}
}

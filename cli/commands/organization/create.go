package organization

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
	Description string `survey:"description"`
	Name        string `survey:"name"`
}

// CreateCommand adds command that allows users to create new organizations
func CreateCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create NAME",
		Short:        "create new organization",
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
					opts.Name = args[0]
				}
			}

			org := opts.toOrganization()
			if err := org.Validate(); err != nil {
				if !isInteractive {
					cmd.SilenceUsage = false
				}
				return err
			}

			if err := cli.Client.CreateOrganization(org); err != nil {
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

	cmd.Flags().StringP("description", "", "", "Description of organization")
	cmd.Flags().StringP("name", "", "", "Name of organization")

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

func (opts *createOpts) toOrganization() *types.Organization {
	return &types.Organization{
		Description: opts.Description,
		Name:        opts.Name,
	}
}

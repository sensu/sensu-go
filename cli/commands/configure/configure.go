package configure

import (
	"errors"
	"fmt"

	"github.com/AlecAivazis/survey"
	"github.com/sensu/sensu-go/cli"
	config "github.com/sensu/sensu-go/cli/client/config"
	hooks "github.com/sensu/sensu-go/cli/commands/hooks"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type configureAnswers struct {
	URL          string `survey:"url"`
	Username     string `survey:"username"`
	Password     string
	Environment  string `survey:"environment"`
	Format       string `survey:"format"`
	Organization string `survey:"organization"`
}

// Command defines new configuration command
func Command(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "configure",
		Short:        "Initialize sensuctl configuration",
		SilenceUsage: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			flags := cmd.Flags()
			nonInteractive, _ := flags.GetBool("non-interactive")
			if nonInteractive {
				// Mark flags are required for bash-completions
				_ = cmd.MarkFlagRequired("username")
				_ = cmd.MarkFlagRequired("password")
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			flags := cmd.Flags()

			nonInteractive, err := flags.GetBool("non-interactive")
			if err != nil {
				return err
			}

			answers := &configureAnswers{}

			if nonInteractive {
				answers.withFlags(flags)
			} else {
				if err = answers.administerQuestionnaire(cli.Config); err != nil {
					return err
				}
			}

			// Write new API URL to disk
			if err = cli.Config.SaveAPIUrl(answers.URL); err != nil {
				fmt.Fprintln(cmd.OutOrStderr())
				return fmt.Errorf(
					"unable to write new configuration file with error: %s",
					err,
				)
			}

			// Authenticate
			tokens, err := cli.Client.CreateAccessToken(
				answers.URL, answers.Username, answers.Password,
			)
			if err != nil {
				fmt.Fprintln(cmd.OutOrStderr())
				return fmt.Errorf("unable to authenticate with error: %s", err)
			} else if tokens == nil {
				fmt.Fprintln(cmd.OutOrStderr())
				return fmt.Errorf("bad username or password")
			}

			// Write new credentials to disk
			if err = cli.Config.SaveTokens(tokens); err != nil {
				fmt.Fprintln(cmd.OutOrStderr())
				return fmt.Errorf(
					"unable to write new configuration file with error: %s",
					err,
				)
			}

			if err = cli.Config.SaveEnvironment(answers.Environment); err != nil {
				fmt.Fprintln(cmd.OutOrStderr())
				return fmt.Errorf(
					"unable to write new configuration file with error: %s",
					err,
				)
			}

			// Write CLI preferences to disk
			if err = cli.Config.SaveFormat(answers.Format); err != nil {
				fmt.Fprintln(cmd.OutOrStderr())
				return fmt.Errorf(
					"unable to write new configuration file with error: %s",
					err,
				)
			}

			if err = cli.Config.SaveOrganization(answers.Organization); err != nil {
				fmt.Fprintln(cmd.OutOrStderr())
				return fmt.Errorf(
					"unable to write new configuration file with error: %s",
					err,
				)
			}

			return nil
		},
		Annotations: map[string]string{
			// We want to be able to run this command regardless of whether the CLI
			// has been configured.
			hooks.ConfigurationRequirement: hooks.ConfigurationNotRequired,
		},
	}

	_ = cmd.Flags().BoolP("non-interactive", "n", false, "do not administer interactive questionnaire")
	_ = cmd.Flags().StringP("url", "", cli.Config.APIUrl(), "the sensu backend url")
	_ = cmd.Flags().StringP("username", "", "", "username")
	_ = cmd.Flags().StringP("password", "", "", "password")
	_ = cmd.Flags().StringP("environment", "", cli.Config.Environment(), "environment")
	_ = cmd.Flags().StringP("format", "", cli.Config.Format(), "preferred output format")
	_ = cmd.Flags().StringP("organization", "", cli.Config.Organization(), "organization")

	return cmd
}

func (answers *configureAnswers) administerQuestionnaire(c config.Config) error {
	qs := []*survey.Question{
		askForURL(c),
		askForUsername(),
		askForPassword(),
		askForOrganization(c),
		askForEnvironment(c),
		askForDefaultFormat(c),
	}

	return survey.Ask(qs, answers)
}

func (answers *configureAnswers) withFlags(flags *pflag.FlagSet) {
	answers.URL, _ = flags.GetString("url")
	answers.Username, _ = flags.GetString("username")
	answers.Password, _ = flags.GetString("password")
	answers.Environment, _ = flags.GetString("environment")
	answers.Format, _ = flags.GetString("format")
	answers.Organization, _ = flags.GetString("organization")
}

func askForURL(c config.Config) *survey.Question {
	url := c.APIUrl()

	return &survey.Question{
		Name: "url",
		Prompt: &survey.Input{
			Message: "Sensu Backend URL:",
			Default: url,
		},
	}
}

func askForUsername() *survey.Question {
	return &survey.Question{
		Name: "username",
		Prompt: &survey.Input{
			Message: "Username:",
			Default: "",
		},
	}
}

func askForPassword() *survey.Question {
	return &survey.Question{
		Name:   "password",
		Prompt: &survey.Password{Message: "Password:"},
	}
}

func askForDefaultFormat(c config.Config) *survey.Question {
	format := c.Format()

	return &survey.Question{
		Name: "format",
		Prompt: &survey.Select{
			Message: "Preferred output format:",
			Options: []string{"none", "json"},
			Default: format,
		},
	}
}

func askForEnvironment(c config.Config) *survey.Question {
	env := c.Environment()

	return &survey.Question{
		Name: "environment",
		Prompt: &survey.Input{
			Message: "Environment:",
			Default: env,
		},
	}
}

func askForOrganization(c config.Config) *survey.Question {
	organization := c.Organization()

	return &survey.Question{
		Name: "organization",
		Prompt: &survey.Input{
			Message: "Organization:",
			Default: organization,
		},
	}
}

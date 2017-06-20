package configure

import (
	"fmt"

	"github.com/AlecAivazis/survey"
	"github.com/sensu/sensu-go/cli"
	config "github.com/sensu/sensu-go/cli/client/config"
	hooks "github.com/sensu/sensu-go/cli/commands/hooks"
	"github.com/spf13/cobra"
)

const (
	defaultFormat       = "none"
	defaultOrganization = "default"
)

type answers struct {
	URL          string `survey:"url"`
	Username     string `survey:"username"`
	Password     string
	Format       string `survey:"format"`
	Organization string `survey:"organization"`
}

// Command defines new configuration command
func Command(cli *cli.SensuCli) *cobra.Command {
	return &cobra.Command{
		Use:          "configure",
		Short:        "Configure Sensu CLI options",
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			// Get new values via interactive questions
			configValues, err := gatherConfigValues(cli.Config)
			if err != nil {
				fmt.Fprintln(cmd.OutOrStderr(), err)
				return
			}

			// Write new API URL to disk
			if err = cli.Config.SaveAPIUrl(configValues.URL); err != nil {
				fmt.Fprintf(
					cmd.OutOrStderr(),
					"Unable to write new configuration file with error: %s.\n",
					err,
				)
				return
			}

			// Authenticate
			tokens, err := cli.Client.CreateAccessToken(
				configValues.URL, configValues.Username, configValues.Password,
			)
			if err != nil {
				fmt.Fprintf(
					cmd.OutOrStderr(),
					"Unable to authenticate with error: %s.\n",
					err,
				)
				return
			} else if tokens == nil {
				fmt.Fprintln(cmd.OutOrStderr(), "Bad username or password.")
				return
			}

			// Write new credentials to disk
			if err = cli.Config.SaveTokens(tokens); err != nil {
				fmt.Fprintf(
					cmd.OutOrStderr(),
					"Unable to write new configuration file with error: %s\n",
					err,
				)
			}

			// Write CLI preferences to disk
			if err = cli.Config.SaveFormat(configValues.Format); err != nil {
				fmt.Fprintf(
					cmd.OutOrStderr(),
					"Unable to write new configuration file with error: %s\n",
					err,
				)
			}

			if err = cli.Config.SaveOrganization(configValues.Organization); err != nil {
				fmt.Fprintf(
					cmd.OutOrStderr(),
					"Unable to write new configuration file with error: %s\n",
					err,
				)
			}

			return
		},
		Annotations: map[string]string{
			// We want to be able to run this command regardless of whether the CLI
			// has been configured.
			hooks.ConfigurationRequirement: hooks.ConfigurationNotRequired,
		},
	}
}

func gatherConfigValues(c config.Config) (*answers, error) {
	qs := []*survey.Question{
		askForURL(c),
		askForUsername(),
		askForPassword(),
		askForOrganization(c),
		askForDefaultFormat(c),
	}

	res := &answers{}
	err := survey.Ask(qs, res)
	return res, err
}

func askForURL(c config.Config) *survey.Question {
	url := c.APIUrl()

	return &survey.Question{
		Name: "url",
		Prompt: &survey.Input{
			Message: "Sensu Base URL:",
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
	if format == "" {
		format = defaultFormat
	}

	return &survey.Question{
		Name: "format",
		Prompt: &survey.Select{
			Message: "Preferred output format:",
			Options: []string{"none", "json", "yaml"},
			Default: format,
		},
	}
}

func askForOrganization(c config.Config) *survey.Question {
	organization := c.Organization()
	if organization == "" {
		organization = defaultOrganization
	}

	return &survey.Question{
		Name: "organization",
		Prompt: &survey.Input{
			Message: "Organization:",
			Default: organization,
		},
	}
}

package configure

import (
	"fmt"

	"github.com/AlecAivazis/survey"
	"github.com/sensu/sensu-go/cli"
	clientconfig "github.com/sensu/sensu-go/cli/client/config"
	hooks "github.com/sensu/sensu-go/cli/commands/hooks"
	"github.com/spf13/cobra"
)

type answers struct {
	URL      string `survey:"url"`
	UserID   string `survey:"userid"`
	Password string
	Output   string
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
			if err = cli.Config.WriteURL(configValues.URL); err != nil {
				fmt.Fprintf(
					cmd.OutOrStderr(),
					"Unable to write new configuration file with error: %s.\n",
					err,
				)
				return
			}

			// Authenticate
			token, err := cli.Client.CreateAccessToken(configValues.UserID, configValues.Password)
			if err != nil {
				fmt.Fprintf(
					cmd.OutOrStderr(),
					"Unable to authenticate with error: %s.\n",
					err,
				)
				return
			} else if token == nil {
				fmt.Fprintln(cmd.OutOrStderr(), "Bad username or password.")
				return
			}

			// Write new credentials to disk
			if err = cli.Config.WriteCredentials(token); err != nil {
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

func gatherConfigValues(config clientconfig.Config) (*answers, error) {
	qs := []*survey.Question{
		askForURL(config),
		askForUsername(),
		askForPassword(),
		askForDefaultOutput(),
	}

	res := &answers{}
	err := survey.Ask(qs, res)
	return res, err
}

func askForURL(config clientconfig.Config) *survey.Question {
	url := config.GetString("api-url")

	return &survey.Question{
		Name:   "url",
		Prompt: &survey.Input{"Sensu Base URL:", url},
	}
}

func askForUsername() *survey.Question {
	return &survey.Question{
		Name:   "userid",
		Prompt: &survey.Input{"Email:", ""},
	}
}

func askForPassword() *survey.Question {
	return &survey.Question{
		Name:   "password",
		Prompt: &survey.Password{Message: "Password:"},
	}
}

func askForDefaultOutput() *survey.Question {
	return &survey.Question{
		Name: "output",
		Prompt: &survey.Select{
			Message: "Preferred output:",
			Options: []string{"none", "json", "yaml"},
			Default: "none",
		},
	}
}

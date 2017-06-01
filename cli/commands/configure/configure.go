package configure

import (
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
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get new values via interactive questions
			_, err := gatherConfigValues(cli.Config)
			if err != nil {
				return err
			}

			// TODO: Authenticate.
			// TODO: Write new configuration file.

			return nil
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

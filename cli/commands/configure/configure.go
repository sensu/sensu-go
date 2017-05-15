package configure

import (
	"fmt"
	"os"
	"path"

	"github.com/AlecAivazis/survey"
	toml "github.com/pelletier/go-toml"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client"
	hooks "github.com/sensu/sensu-go/cli/commands/hooks"
	"github.com/spf13/cobra"
)

type answers struct {
	URL    string `survey:"url"`
	UserID string `survey:"userid"`
	Secret string
	Output string
}

// Command defines new configuration command
func Command(cli *cli.SensuCli) *cobra.Command {
	return &cobra.Command{
		Use:          "configure",
		Short:        "Configure Sensu CLI options",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			config := emptyTomlTree()

			// Read configuration file
			if _, err := os.Stat(client.ConfigFilePath); err == nil {
				config, err = toml.LoadFile(client.ConfigFilePath)
				if err != nil {
					fmt.Fprintln(os.Stderr, "Error loading config:")
					return err
				}
			} else {
				// Ensure that the path to the configuration exists
				os.MkdirAll(path.Dir(client.ConfigFilePath), 0755)
			}

			// Get the configuation values for the specified profile
			profileKey := cli.Config.GetString("profile")
			profile, ok := config.Get(profileKey).(*toml.TomlTree)
			if !ok {
				profile = emptyTomlTree()
			}

			// Get new values via interactive questions
			v, err := gatherConfigValues(profile)
			if err != nil {
				return err
			}

			// Update profile
			profile.Set("url", v.URL)
			profile.Set("userid", v.UserID)
			profile.Set("secret", v.Secret)
			config.Set(profileKey, profile)

			// Write config
			writeNewConfig(config)

			return nil
		},
		Annotations: map[string]string{
			// We want to be able to run this command regardless of whether the CLI
			// has been configured.
			hooks.ConfigurationRequirement: hooks.ConfigurationNotRequired,
		},
	}
}

func gatherConfigValues(config *toml.TomlTree) (*answers, error) {
	qs := []*survey.Question{
		askForURL(config),
		askForUsername(config),
		askForSecret(),
		askForDefaultOutput(),
	}

	res := &answers{}
	err := survey.Ask(qs, res)
	return res, err
}

func askForURL(config *toml.TomlTree) *survey.Question {
	url, _ := config.Get("url").(string)

	return &survey.Question{
		Name:   "url",
		Prompt: &survey.Input{"Sensu Base URL:", url},
	}
}

func askForUsername(config *toml.TomlTree) *survey.Question {
	userid, _ := config.Get("userid").(string)

	return &survey.Question{
		Name:   "userid",
		Prompt: &survey.Input{"Email:", userid},
	}
}

func askForSecret() *survey.Question {
	return &survey.Question{
		Name:   "secret",
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

func writeNewConfig(config *toml.TomlTree) error {
	f, err := os.Create(client.ConfigFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = config.WriteTo(f)
	return err
}

func emptyTomlTree() *toml.TomlTree {
	empty, _ := toml.TreeFromMap(make(map[string]interface{}))
	return empty
}

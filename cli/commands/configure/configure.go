package configure

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey"
	toml "github.com/pelletier/go-toml"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client"
	"github.com/spf13/cobra"
)

type Answers struct {
	URL    string `survey 'url'`
	UserID string `survey 'userid'`
	Secret string
	Output string
}

func Command(cli *cli.SensuCli) *cobra.Command {
	return &cobra.Command{
		Use:   "configure",
		Short: "Configure Sensu CLI options",
		RunE: func(cmd *cobra.Command, args []string) error {
			config := emptyTomlTree()

			// Read configuration file
			// TODO(james) handle case where file exists but is invalid
			if _, err := os.Stat(client.ConfigFilePath); err == nil {
				config, err = toml.LoadFile(client.ConfigFilePath)
				if err != nil {
					fmt.Fprintln(os.Stderr, "Error loading config:")
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
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
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			fmt.Fprintln(os.Stderr, v)

			// Update profile
			profile.Set("url", v.URL)
			profile.Set("userid", v.UserID)
			profile.Set("secret", v.Secret)
			config.Set(profileKey, profile)

			// Write config
			writeNewConfig(config)

			return nil
		},
	}
}

func gatherConfigValues(config *toml.TomlTree) (*Answers, error) {
	qs := []*survey.Question{
		AskForURL(config),
		AskForUsername(config),
		AskForSecret(),
		AskForDefaultOutput(),
	}

	answers := &Answers{}
	err := survey.Ask(qs, answers)
	return answers, err
}

func AskForURL(config *toml.TomlTree) *survey.Question {
	url, _ := config.Get("url").(string)

	return &survey.Question{
		Name:   "url",
		Prompt: &survey.Input{"Sensu Base URL:", url},
	}
}

func AskForUsername(config *toml.TomlTree) *survey.Question {
	userid, _ := config.Get("userid").(string)

	return &survey.Question{
		Name:   "userid",
		Prompt: &survey.Input{"Email:", userid},
	}
}

func AskForSecret() *survey.Question {
	return &survey.Question{
		Name:   "secret",
		Prompt: &survey.Password{Message: "Password:"},
	}
}

func AskForDefaultOutput() *survey.Question {
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

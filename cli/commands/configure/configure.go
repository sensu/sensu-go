package configure

import (
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	toml "github.com/pelletier/go-toml"
	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
	"gopkg.in/AlecAivazis/survey.v1"
)

var configFile string

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
			// Read configuration file
			// TODO(james) handle case where file exists but is invalid
			c, err := toml.LoadFile(configFile)
			if err != nil {
				c, _ = toml.TreeFromMap(make(map[string]interface{}))
			}

			// Get the configuation values for the specified profile
			profileKey := cli.Config.GetString("profile")
			profile, ok := c.Get(profileKey).(*toml.TomlTree)
			if !ok {
				profile, _ = toml.TreeFromMap(make(map[string]interface{}))
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
			c.Set(profileKey, profile)

			// Write config
			writeNewConfig(c)

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
		Name:     "url",
		Prompt:   &survey.Input{"Sensu Base URL:", url},
		Validate: survey.Required,
	}
}

func AskForUsername(config *toml.TomlTree) *survey.Question {
	userid, _ := config.Get("userid").(string)

	return &survey.Question{
		Name:     "userid",
		Prompt:   &survey.Input{"Email:", userid},
		Validate: survey.Required,
	}
}

func AskForSecret() *survey.Question {
	return &survey.Question{
		Name:     "secret",
		Prompt:   &survey.Password{Message: "Password:"},
		Validate: survey.Required,
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
	f, err := os.Create(configFile)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = config.WriteTo(f)
	return err
}

func init() {
	h, _ := homedir.Dir()
	// TODO(james) move configuration into it's own package
	configFile = filepath.Join(h, ".config", "sensu", "profiles.toml")
}

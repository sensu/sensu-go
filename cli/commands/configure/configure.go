package configure

import (
	"errors"
	"fmt"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client/config"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/commands/hooks"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	FlagFormat                = "format"
	FlagInsecureSkipTlsVerify = "insecure-skip-tls-verify"
	FlagNamespace             = "namespace"
	FlagNonInteractive        = "non-interactive"
	FlagPassword              = "password"
	FlagTimeout               = "timeout"
	FlagTrustedCaFile         = "trusted-ca-file"
	FlagUrl                   = "url"
	FlagUsername              = "username"
)

type Answers struct {
	URL                   string `survey:"url"`
	Username              string `survey:"username"`
	Password              string
	Format                string `survey:"format"`
	Namespace             string `survey:"namespace"`
	InsecureSkipTLSVerify bool
	Timeout               time.Duration
	TrustedCAFile         string
}

// Command defines new configuration command
func Command(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "configure",
		Short:        "Initialize sensuctl configuration",
		SilenceUsage: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			flags := cmd.Flags()
			v, err := helpers.InitViper(flags)
			if err != nil {
				return
			}

			nonInteractive := v.GetBool(FlagNonInteractive)
			if nonInteractive {
				username := v.GetString(FlagUsername)
				password := v.GetString(FlagPassword)
				if username == "" || password == "" {
					// Mark flags are required for bash-completions when they're missing
					_ = cmd.MarkFlagRequired(FlagUsername)
					_ = cmd.MarkFlagRequired(FlagPassword)
				}
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			flags := cmd.Flags()
			v, err := helpers.InitViper(flags)
			if err != nil {
				return err
			}

			nonInteractive := v.GetBool(FlagNonInteractive)

			answers := &Answers{}
			if nonInteractive {
				answers.WithFlags(v)
			} else {
				if err := answers.AdministerQuestionnaire(cli.Config); err != nil {
					return err
				}
			}

			// First save the API URL
			if err := SaveAPIURL(cli, answers); err != nil {
				_, _ = fmt.Fprintln(cmd.OutOrStderr())
				return err
			}

			if err := Authenticate(cli, answers); err != nil {
				_, _ = fmt.Fprintln(cmd.OutOrStderr())
				return err
			}

			return SaveConfig(cli, answers, v)
		},
		Annotations: map[string]string{
			// We want to be able to run this command regardless of whether the CLI
			// has been configured.
			hooks.ConfigurationRequirement: hooks.ConfigurationNotRequired,
		},
	}

	AddFlags(cli, cmd)

	return cmd
}

func AddFlags(cli *cli.SensuCli, cmd *cobra.Command) {
	_ = cmd.Flags().BoolP(FlagNonInteractive, "n", false, "do not administer interactive questionnaire")
	_ = cmd.Flags().StringP(FlagUrl, "", cli.Config.APIUrl(), "the sensu backend url")
	_ = cmd.Flags().StringP(FlagUsername, "", "", "username")
	_ = cmd.Flags().StringP(FlagPassword, "", "", "password")
	_ = cmd.Flags().StringP(FlagFormat, "", cli.Config.Format(), "preferred output format")
	_ = cmd.Flags().StringP(FlagNamespace, "", cli.Config.Namespace(), "namespace")
	_ = cmd.Flags().DurationP(FlagTimeout, "", cli.Config.Timeout(), "timeout when communicating with backend url")
}

func (answers *Answers) AdministerQuestionnaire(c config.Config) error {
	qs := []*survey.Question{
		AskForURL(c),
		AskForUsername(),
		AskForPassword(),
		AskForNamespace(c),
		AskForDefaultFormat(c),
	}

	return survey.Ask(qs, answers)
}

func (answers *Answers) WithFlags(v *viper.Viper) {
	answers.URL = v.GetString(FlagUrl)
	answers.Username = v.GetString(FlagUsername)
	answers.Password = v.GetString(FlagPassword)
	answers.Format = v.GetString(FlagFormat)
	answers.Namespace = v.GetString(FlagNamespace)
	answers.Timeout = v.GetDuration(FlagTimeout)
}

func AskForURL(c config.Config) *survey.Question {
	url := c.APIUrl()

	return &survey.Question{
		Name: "url",
		Prompt: &survey.Input{
			Message: "Sensu Backend API URL:",
			Default: url,
		},
	}
}

func AskForUsername() *survey.Question {
	return &survey.Question{
		Name: "username",
		Prompt: &survey.Input{
			Message: "Username:",
			Default: "",
		},
	}
}

func AskForPassword() *survey.Question {
	return &survey.Question{
		Name:   "password",
		Prompt: &survey.Password{Message: "Password:"},
	}
}

func AskForDefaultFormat(c config.Config) *survey.Question {
	format := c.Format()

	return &survey.Question{
		Name: "format",
		Prompt: &survey.Select{
			Message: "Preferred output format:",
			Options: []string{
				config.FormatTabular,
				config.FormatYAML,
				config.FormatJSON,
			},
			Default: format,
		},
	}
}

func AskForNamespace(c config.Config) *survey.Question {
	namespace := c.Namespace()

	return &survey.Question{
		Name: "namespace",
		Prompt: &survey.Input{
			Message: "Namespace:",
			Default: namespace,
		},
	}
}

func Authenticate(cli *cli.SensuCli, answers *Answers) error {
	// Authenticate
	tokens, err := cli.Client.CreateAccessToken(
		answers.URL, answers.Username, answers.Password,
	)
	if err != nil {
		return fmt.Errorf("unable to authenticate with error: %s", err)
	} else if tokens == nil {
		return fmt.Errorf("bad username or password")
	}

	// Write new credentials to disk
	if err = cli.Config.SaveTokens(tokens); err != nil {
		return fmt.Errorf(
			"unable to write new configuration file with error: %s",
			err,
		)
	}

	return nil
}

// SaveAPIURL saves the backend API URL
func SaveAPIURL(cli *cli.SensuCli, answers *Answers) error {
	// Write new API URL to disk
	if err := cli.Config.SaveAPIUrl(answers.URL); err != nil {
		return fmt.Errorf(
			"unable to write new configuration file with error: %s",
			err,
		)
	}
	return nil
}

// SaveConfig writes to disk the user preferences
func SaveConfig(cli *cli.SensuCli, answers *Answers, v *viper.Viper) error {
	// Write CLI preferences to disk
	if err := cli.Config.SaveFormat(answers.Format); err != nil {
		return fmt.Errorf(
			"unable to write new configuration file with error: %s",
			err,
		)
	}

	if err := cli.Config.SaveNamespace(answers.Namespace); err != nil {
		return fmt.Errorf(
			"unable to write new configuration file with error: %s",
			err,
		)
	}

	// Write the TLS preferences to disk
	skipTlsValue := v.GetBool(FlagInsecureSkipTlsVerify)
	if err := cli.Config.SaveInsecureSkipTLSVerify(skipTlsValue); err != nil {
		return fmt.Errorf(
			"unable to write new configuration file with error: %s",
			err,
		)
	}

	caFileValue := v.GetString(FlagTrustedCaFile)
	if err := cli.Config.SaveTrustedCAFile(caFileValue); err != nil {
		return fmt.Errorf(
			"unable to write new configuration file with error: %s",
			err,
		)
	}

	timeoutValue := v.GetString(FlagTimeout)
	duration, err := time.ParseDuration(timeoutValue)
	if err != nil {
		return fmt.Errorf(
			"unable to parse timeout with error: %s",
			err,
		)
	}
	if err = cli.Config.SaveTimeout(duration); err != nil {
		return fmt.Errorf(
			"unable to write new configuration file with error: %s",
			err,
		)
	}

	return nil
}

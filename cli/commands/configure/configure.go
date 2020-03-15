package configure

import (
	"errors"
	"fmt"
	"time"

	"github.com/AlecAivazis/survey"
	"github.com/sensu/sensu-go/cli"
	config "github.com/sensu/sensu-go/cli/client/config"
	hooks "github.com/sensu/sensu-go/cli/commands/hooks"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type configureAnswers struct {
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

			// Write CLI preferences to disk
			if err = cli.Config.SaveFormat(answers.Format); err != nil {
				fmt.Fprintln(cmd.OutOrStderr())
				return fmt.Errorf(
					"unable to write new configuration file with error: %s",
					err,
				)
			}

			if err = cli.Config.SaveNamespace(answers.Namespace); err != nil {
				fmt.Fprintln(cmd.OutOrStderr())
				return fmt.Errorf(
					"unable to write new configuration file with error: %s",
					err,
				)
			}

			// Write the TLS preferences to disk
			if value, err := flags.GetBool("insecure-skip-tls-verify"); err == nil {
				if err = cli.Config.SaveInsecureSkipTLSVerify(value); err != nil {
					fmt.Fprintln(cmd.OutOrStderr())
					return fmt.Errorf(
						"unable to write new configuration file with error: %s",
						err,
					)
				}
			}
			if value, err := flags.GetString("trusted-ca-file"); err == nil {
				if err = cli.Config.SaveTrustedCAFile(value); err != nil {
					fmt.Fprintln(cmd.OutOrStderr())
					return fmt.Errorf(
						"unable to write new configuration file with error: %s",
						err,
					)
				}
			}

			if value, err := flags.GetString("timeout"); err == nil {
				duration, err := time.ParseDuration(value)
				if err != nil {
					fmt.Fprintln(cmd.OutOrStderr())
					return fmt.Errorf(
						"unable to parse timeout with error: %s",
						err,
					)
				}
				if err = cli.Config.SaveTimeout(duration); err != nil {
					fmt.Fprintln(cmd.OutOrStderr())
					return fmt.Errorf(
						"unable to write new configuration file with error: %s",
						err,
					)
				}
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
	_ = cmd.Flags().StringP("format", "", cli.Config.Format(), "preferred output format")
	_ = cmd.Flags().StringP("namespace", "", cli.Config.Namespace(), "namespace")
	_ = cmd.Flags().DurationP("timeout", "", cli.Config.Timeout(), "timeout when communicating with backend url")

	return cmd
}

func (answers *configureAnswers) administerQuestionnaire(c config.Config) error {
	qs := []*survey.Question{
		askForURL(c),
		askForUsername(),
		askForPassword(),
		askForNamespace(c),
		askForDefaultFormat(c),
	}

	return survey.Ask(qs, answers)
}

func (answers *configureAnswers) withFlags(flags *pflag.FlagSet) {
	answers.URL, _ = flags.GetString("url")
	answers.Username, _ = flags.GetString("username")
	answers.Password, _ = flags.GetString("password")
	answers.Format, _ = flags.GetString("format")
	answers.Namespace, _ = flags.GetString("namespace")
	answers.Timeout, _ = flags.GetDuration("timeout")
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
			Options: []string{
				config.FormatTabular,
				config.FormatYAML,
				config.FormatWrappedJSON,
				config.FormatJSON,
			},
			Default: format,
		},
	}
}

func askForNamespace(c config.Config) *survey.Question {
	namespace := c.Namespace()

	return &survey.Question{
		Name: "namespace",
		Prompt: &survey.Input{
			Message: "Namespace:",
			Default: namespace,
		},
	}
}

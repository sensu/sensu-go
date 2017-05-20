package check

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type checkOpts struct {
	Name          string `survey:"name"`
	Command       string `survey:"command"`
	Interval      string `survey:"interval"`
	Subscriptions string `survey:"subscriptions"`
	Handlers      string `survey:"handlers"`
	Dependencies  string `survey:"dependencies"`
}

const (
	intervalDefault = "60"
)

// CreateCommand adds command that allows user to create new checks
func CreateCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create NAME",
		Short:        "create new checks",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			flags := cmd.Flags()
			isInteractive := flags.NFlag() == 0
			opts := &checkOpts{}

			if isInteractive {
				opts.administerQuestionnaire()
			} else {
				opts.withFlags(flags)
				if len(args) > 0 {
					opts.Name = args[0]
				}
			}

			check := buildCheck(opts, cli.Client)
			if err := check.Validate(); err != nil {
				if !isInteractive {
					cmd.SilenceUsage = false
				}
				return err
			}

			if err := cli.Client.CreateCheck(check); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "OK")
			return nil
		},
	}

	cmd.Flags().StringP("command", "c", "", "the command the check should run")
	cmd.Flags().StringP("interval", "i", intervalDefault, "interval, in second, at which the check is run")
	cmd.Flags().StringSliceP("subscriptions", "s", []string{}, "comma separated list of subscribers")
	cmd.Flags().StringSlice("handlers", []string{}, "comma separated list of handlers")
	cmd.Flags().StringSliceP("runtime-dependency", "d", []string{}, "asset this check depends on")
	cmd.Flags().StringSlice("runtime-dependency-url", []string{}, "URL of asset this check depends on")

	// Mark flags are required for bash-completions
	cmd.MarkFlagRequired("command")
	cmd.MarkFlagRequired("interval")
	cmd.MarkFlagRequired("subscription")

	return cmd
}

func (opts *checkOpts) withFlags(flags *pflag.FlagSet) {
	opts.Command, _ = flags.GetString("command")
	opts.Interval, _ = flags.GetString("interval")

	subscriptions, _ := flags.GetStringSlice("subscriptions")
	opts.Subscriptions = strings.Join(subscriptions, ",")

	handlers, _ := flags.GetStringSlice("handlers")
	opts.Handlers = strings.Join(handlers, ",")

	dependencies, _ := flags.GetStringSlice("runtime-dependency")
	opts.Dependencies = strings.Join(dependencies, ",")
}

func (opts *checkOpts) administerQuestionnaire() {
	var qs = []*survey.Question{
		{
			Name:     "name",
			Prompt:   &survey.Input{"Check Name:", ""},
			Validate: survey.Required,
		},
		{
			Name:     "command",
			Prompt:   &survey.Input{"Command:", ""},
			Validate: survey.Required,
		},
		{
			Name: "interval",
			Prompt: &survey.Input{
				"Interval:",
				intervalDefault,
			},
		},
		{
			Name:     "subscriptions",
			Prompt:   &survey.Input{"Subscriptions:", ""},
			Validate: survey.Required,
		},
		{
			Name:     "handlers",
			Prompt:   &survey.Input{"Handlers:", ""},
			Validate: survey.Required,
		},
		{
			Name:     "dependencies",
			Prompt:   &survey.Input{"Runtime Dependencies:", ""},
			Validate: survey.Required,
		},
	}

	survey.Ask(qs, opts)
}

func buildCheck(opts *checkOpts, client client.APIClient) *types.Check {
	interval, _ := strconv.Atoi(opts.Interval)
	check := &types.Check{
		Name:                opts.Name,
		Interval:            interval,
		Command:             opts.Command,
		Subscriptions:       strings.Split(opts.Subscriptions, ","),
		Handlers:            strings.Split(opts.Handlers, ","),
		RuntimeDependencies: []types.Asset{},
	}

	if len(opts.Dependencies) > 0 {
		assets, _ := client.ListAssets()
		dependencies := strings.Split(opts.Dependencies, ",")

		for _, asset := range assets {
			for _, givenName := range dependencies {
				if asset.Name == givenName {
					check.RuntimeDependencies = append(check.RuntimeDependencies, asset)
					break
				}
			}
		}
	}

	return check
}

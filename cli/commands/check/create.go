package check

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey"
	"github.com/sensu/sensu-go/cli"
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
}

const (
	intervalDefault = "60"
)

// CreateCommand adds command that allows user to create new checks
func CreateCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [COMMAND]",
		Short: "create new checks",
		RunE: func(cmd *cobra.Command, args []string) error {
			flags := cmd.Flags()
			isInteractive := flags.NFlag() == 0
			opts := &checkOpts{}

			if isInteractive {
				opts.administerQuestionnaire()
			} else {
				opts.withFlags(flags)
				if len(args) > 0 {
					opts.Command = args[0]
				}
			}

			check := opts.toCheck()
			if err := check.Validate(); err != nil {
				if isInteractive {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				} else {
					return err
				}
			}

			err := cli.Client.CreateCheck(check)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			return nil
		},
	}

	cmd.Flags().StringP("name", "n", "", "name of the check")
	cmd.Flags().StringP("interval", "i", intervalDefault, "interval, in second, at which the check is run")
	cmd.Flags().StringP("subscriptions", "s", "", "comma separated list of subscribers")
	cmd.Flags().String("handlers", "", "comma separated list of handlers")

	return cmd
}

func (opts *checkOpts) withFlags(flags *pflag.FlagSet) {
	opts.Name, _ = flags.GetString("name")
	opts.Interval, _ = flags.GetString("interval")
	opts.Subscriptions, _ = flags.GetString("subscriptions")
	opts.Handlers, _ = flags.GetString("handlers")
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
	}

	_ = survey.Ask(qs, opts)
}

func (opts *checkOpts) toCheck() *types.Check {
	interval, _ := strconv.Atoi(opts.Interval)

	return &types.Check{
		Name:          opts.Name,
		Interval:      interval,
		Command:       opts.Command,
		Subscriptions: strings.Split(opts.Subscriptions, ","),
		Handlers:      strings.Split(opts.Handlers, ","),
	}
}

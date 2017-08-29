package check

import (
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/pflag"
)

const (
	intervalDefault = "60"
)

type checkOpts struct {
	Name          string `survey:"name"`
	Command       string `survey:"command"`
	Interval      string `survey:"interval"`
	Subscriptions string `survey:"subscriptions"`
	Handlers      string `survey:"handlers"`
	RuntimeAssets string `survey:"assets"`
	Env           string
	Org           string
}

func newCheckOpts() *checkOpts {
	opts := checkOpts{}
	opts.Interval = intervalDefault
	return &opts
}

func (opts *checkOpts) withCheck(check *types.CheckConfig) {
	opts.Name = check.Name
	opts.Org = check.Organization
	opts.Env = check.Environment
	opts.Command = check.Command
	opts.Interval = strconv.Itoa(int(check.Interval))
	opts.Subscriptions = strings.Join(check.Subscriptions, ",")
	opts.Handlers = strings.Join(check.Handlers, ",")
	opts.RuntimeAssets = strings.Join(check.RuntimeAssets, ",")
}

func (opts *checkOpts) withFlags(flags *pflag.FlagSet) {
	opts.Command, _ = flags.GetString("command")
	opts.Interval, _ = flags.GetString("interval")
	opts.Subscriptions, _ = flags.GetString("subscriptions")
	opts.Handlers, _ = flags.GetString("handlers")
	opts.RuntimeAssets, _ = flags.GetString("runtime-assets")
}

func (opts *checkOpts) administerQuestionnaire(editing bool) {
	var qs = []*survey.Question{}

	if !editing {
		qs = append(qs, []*survey.Question{
			{
				Name:     "name",
				Prompt:   &survey.Input{"Check Name:", ""},
				Validate: survey.Required,
			},
			{
				Name: "org",
				Prompt: &survey.Input{
					"Organization:",
					opts.Org,
				},
				Validate: survey.Required,
			},
			{
				Name: "env",
				Prompt: &survey.Input{
					"Environment:",
					opts.Env,
				},
				Validate: survey.Required,
			},
		}...)
	}

	qs = append(qs, []*survey.Question{
		{
			Name: "command",
			Prompt: &survey.Input{
				"Command:",
				opts.Command,
			},
			Validate: survey.Required,
		},
		{
			Name: "interval",
			Prompt: &survey.Input{
				"Interval:",
				opts.Interval,
			},
		},
		{
			Name: "subscriptions",
			Prompt: &survey.Input{
				"Subscriptions:",
				opts.Subscriptions,
			},
			Validate: survey.Required,
		},
		{
			Name: "handlers",
			Prompt: &survey.Input{
				"Handlers:",
				opts.Handlers,
			},
		},
		{
			Name: "assets",
			Prompt: &survey.Input{
				"Runtime Assets:",
				opts.RuntimeAssets,
			},
		},
	}...)

	survey.Ask(qs, opts)
}

func (opts *checkOpts) Copy(check *types.CheckConfig) {
	interval, _ := strconv.ParseUint(opts.Interval, 10, 32)

	check.Name = opts.Name
	check.Environment = opts.Env
	check.Organization = opts.Org
	check.Interval = uint(interval)
	check.Command = opts.Command
	check.Subscriptions = helpers.SafeSplitCSV(opts.Subscriptions)
	check.Handlers = helpers.SafeSplitCSV(opts.Handlers)
	check.RuntimeAssets = helpers.SafeSplitCSV(opts.RuntimeAssets)
}

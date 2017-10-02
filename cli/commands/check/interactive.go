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
	Publish       bool `survey:"publish"`
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
	opts.Publish, _ = flags.GetBool("publish")
	opts.RuntimeAssets, _ = flags.GetString("runtime-assets")
	opts.Org, _ = flags.GetString("organization")
	opts.Env, _ = flags.GetString("environment")
}

func (opts *checkOpts) administerQuestionnaire(editing bool) {
	var qs = []*survey.Question{}

	if !editing {
		qs = append(qs, []*survey.Question{
			{
				Name: "name",
				Prompt: &survey.Input{
					Message: "Check Name:",
					Default: "",
				},
				Validate: survey.Required,
			},
			{
				Name: "org",
				Prompt: &survey.Input{
					Message: "Organization:",
					Default: opts.Org,
				},
				Validate: survey.Required,
			},
			{
				Name: "env",
				Prompt: &survey.Input{
					Message: "Environment:",
					Default: opts.Env,
				},
				Validate: survey.Required,
			},
		}...)
	}

	qs = append(qs, []*survey.Question{
		{
			Name: "command",
			Prompt: &survey.Input{
				Message: "Command:",
				Default: opts.Command,
			},
			Validate: survey.Required,
		},
		{
			Name: "interval",
			Prompt: &survey.Input{
				Message: "Interval:",
				Default: opts.Interval,
			},
		},
		{
			Name: "subscriptions",
			Prompt: &survey.Input{
				Message: "Subscriptions:",
				Default: opts.Subscriptions,
			},
			Validate: survey.Required,
		},
		{
			Name: "handlers",
			Prompt: &survey.Input{
				Message: "Handlers:",
				Default: opts.Handlers,
			},
		},
		{
			Name: "assets",
			Prompt: &survey.Input{
				Message: "Runtime Assets:",
				Default: opts.RuntimeAssets,
			},
		},
		{
			Name: "publish",
			Prompt: &survey.Confirm{
				Message: "Publish check requests?",
				Default: true,
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
	check.Publish = opts.Publish
}

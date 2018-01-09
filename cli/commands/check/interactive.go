package check

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/pflag"
)

const (
	intervalDefault = "60"
	stdinDefault    = "false"
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
	Publish       string `survey:"publish"`
	ProxyEntityID string `survey:"proxy-entity-id"`
	Stdin         string `survey:"stdin"`
	Timeout       string `survey:"timeout"`
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
	opts.ProxyEntityID = check.ProxyEntityID
	opts.Stdin = stdinDefault
	opts.Timeout = strconv.Itoa(int(check.Timeout))
}

func (opts *checkOpts) withFlags(flags *pflag.FlagSet) {
	opts.Command, _ = flags.GetString("command")
	opts.Interval, _ = flags.GetString("interval")
	opts.Subscriptions, _ = flags.GetString("subscriptions")
	opts.Handlers, _ = flags.GetString("handlers")
	opts.RuntimeAssets, _ = flags.GetString("runtime-assets")
	publishBool, _ := flags.GetBool("publish")
	opts.Publish = strconv.FormatBool(publishBool)
	opts.ProxyEntityID, _ = flags.GetString("proxy-entity-id")
	opts.Stdin, _ = flags.GetString("stdin")
	opts.Timeout, _ = flags.GetString("timeout")

	if org, _ := flags.GetString("organization"); org != "" {
		opts.Org = org
	}
	if env, _ := flags.GetString("environment"); env != "" {
		opts.Env = env
	}
}

func (opts *checkOpts) administerQuestionnaire(editing bool) error {
	var qs = []*survey.Question{}

	if !editing {
		qs = append(qs, []*survey.Question{
			{
				Name: "name",
				Prompt: &survey.Input{
					Message: "Check Name:",
					Default: opts.Name,
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
			Name: "timeout",
			Prompt: &survey.Input{
				Message: "Timeout:",
				Default: opts.Timeout,
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
			Prompt: &survey.Input{
				Message: "Publish:",
				Help:    "If check requests are published for the check. Value must be true or false.",
				Default: "true",
			},
			Validate: func(val interface{}) error {
				if str := val.(string); str != "false" && str != "true" {
					return fmt.Errorf("Please enter either true or false")
				}
				return nil
			},
		},
		{
			Name: "proxy-entity-id",
			Prompt: &survey.Input{
				Message: "Check Proxy Entity ID:",
				Default: opts.ProxyEntityID,
				Help:    "the check's proxy entity id, used to create a proxy entity for an external resource",
			},
		},
		{
			Name: "stdin",
			Prompt: &survey.Input{
				Message: "Check STDIN:",
				Default: opts.Stdin,
				Help:    "If check accepts JSON event data to the check command's stdin. Defaults to false.",
			},
		},
	}...)

	return survey.Ask(qs, opts)
}

func (opts *checkOpts) Copy(check *types.CheckConfig) {
	interval, _ := strconv.ParseUint(opts.Interval, 10, 32)
	stdin, _ := strconv.ParseBool(opts.Stdin)
	timeout, _ := strconv.ParseUint(opts.Timeout, 10, 32)

	check.Name = opts.Name
	check.Environment = opts.Env
	check.Organization = opts.Org
	check.Interval = uint32(interval)
	check.Command = opts.Command
	check.Subscriptions = helpers.SafeSplitCSV(opts.Subscriptions)
	check.Handlers = helpers.SafeSplitCSV(opts.Handlers)
	check.RuntimeAssets = helpers.SafeSplitCSV(opts.RuntimeAssets)
	check.Publish = opts.Publish == "true"
	check.ProxyEntityID = opts.ProxyEntityID
	check.Stdin = stdin
	check.Timeout = uint32(timeout)
}

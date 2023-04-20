package check

import (
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	cron "github.com/robfig/cron/v3"
	v2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/spf13/pflag"
)

const (
	stdinDefault		= "false"
	roundRobinDefault	= "false"
	publishDefault		= "true"
)

type checkOpts struct {
	Name			string	`survey:"name"`
	Command			string	`survey:"command"`
	Interval		string	`survey:"interval"`
	Cron			string	`survey:"cron"`
	Subscriptions		string	`survey:"subscriptions"`
	Handlers		string	`survey:"handlers"`
	RuntimeAssets		string	`survey:"assets"`
	Namespace		string
	Publish			string	`survey:"publish"`
	ProxyEntityName		string	`survey:"proxy-entity-name"`
	Stdin			string	`survey:"stdin"`
	Timeout			string	`survey:"timeout"`
	TTL			string	`survey:"ttl"`
	HighFlapThreshold	string	`survey:"high-flap-threshold"`
	LowFlapThreshold	string	`survey:"low-flap-threshold"`
	OutputMetricFormat	string	`survey:"output-metric-format"`
	OutputMetricHandlers	string	`survey:"output-metric-handlers"`
	RoundRobin		string	`survey:"round-robin"`
}

func newCheckOpts() *checkOpts {
	opts := checkOpts{}
	opts.RoundRobin = roundRobinDefault
	opts.Stdin = stdinDefault
	opts.Publish = publishDefault
	return &opts
}

func (opts *checkOpts) withCheck(check *v2.CheckConfig) {
	opts.Name = check.Name
	opts.Namespace = check.Namespace
	opts.Command = check.Command
	opts.Interval = strconv.Itoa(int(check.Interval))
	opts.Cron = check.Cron
	opts.Subscriptions = strings.Join(check.Subscriptions, ",")
	opts.Handlers = strings.Join(check.Handlers, ",")
	opts.RuntimeAssets = strings.Join(check.RuntimeAssets, ",")
	opts.ProxyEntityName = check.ProxyEntityName
	opts.Stdin = strconv.FormatBool(check.Stdin)
	opts.Timeout = strconv.Itoa(int(check.Timeout))
	opts.HighFlapThreshold = strconv.Itoa(int(check.HighFlapThreshold))
	opts.LowFlapThreshold = strconv.Itoa(int(check.LowFlapThreshold))
	opts.OutputMetricFormat = check.OutputMetricFormat
	opts.OutputMetricHandlers = strings.Join(check.OutputMetricHandlers, ",")
	opts.RoundRobin = strconv.FormatBool(check.RoundRobin)
	opts.Publish = strconv.FormatBool(check.Publish)
}

func (opts *checkOpts) withFlags(flags *pflag.FlagSet) {
	opts.Command, _ = flags.GetString("command")
	opts.Interval, _ = flags.GetString("interval")
	opts.Cron, _ = flags.GetString("cron")
	opts.Subscriptions, _ = flags.GetString("subscriptions")
	opts.Handlers, _ = flags.GetString("handlers")
	opts.RuntimeAssets, _ = flags.GetString("runtime-assets")
	publishBool, _ := flags.GetBool("publish")
	opts.Publish = strconv.FormatBool(publishBool)
	opts.ProxyEntityName, _ = flags.GetString("proxy-entity-name")
	opts.Stdin, _ = flags.GetString("stdin")
	opts.Timeout, _ = flags.GetString("timeout")
	opts.TTL, _ = flags.GetString("ttl")
	opts.HighFlapThreshold, _ = flags.GetString("high-flap-threshold")
	opts.LowFlapThreshold, _ = flags.GetString("low-flap-threshold")
	opts.OutputMetricFormat, _ = flags.GetString("output-metric-format")
	opts.OutputMetricHandlers, _ = flags.GetString("output-metric-handlers")
	roundRobinBool, _ := flags.GetBool("round-robin")
	opts.RoundRobin = strconv.FormatBool(roundRobinBool)

	if namespace := helpers.GetChangedStringValueViper("namespace", flags); namespace != "" {
		opts.Namespace = namespace
	}
}

func (opts *checkOpts) administerQuestionnaire(editing bool) error {
	var qs = []*survey.Question{}

	if !editing {
		qs = append(qs, []*survey.Question{
			{
				Name:	"name",
				Prompt: &survey.Input{
					Message:	"Check Name:",
					Default:	opts.Name,
				},
				Validate:	survey.Required,
			},
			{
				Name:	"namespace",
				Prompt: &survey.Input{
					Message:	"Namespace:",
					Default:	opts.Namespace,
				},
				Validate:	survey.Required,
			},
		}...)
	}

	qs = append(qs, []*survey.Question{
		{
			Name:	"command",
			Prompt: &survey.Input{
				Message:	"Command:",
				Default:	opts.Command,
			},
			Validate:	survey.Required,
		},
		{
			Name:	"interval",
			Prompt: &survey.Input{
				Message:	"Interval:",
				Default:	opts.Interval,
			},
		},
		{
			Name:	"cron",
			Prompt: &survey.Input{
				Message:	"Cron:",
				Help:		"Optional cron schedule which takes precedence over interval. Value must be a valid cron string.",
				Default:	opts.Cron,
			},
			Validate: func(val interface{}) error {
				if value, ok := val.(string); ok && value != "" {
					if _, err := cron.ParseStandard(value); err != nil {
						return err
					}
				}
				return nil
			},
		},
		{
			Name:	"timeout",
			Prompt: &survey.Input{
				Message:	"Timeout:",
				Default:	opts.Timeout,
			},
		},
		{
			Name:	"ttl",
			Prompt: &survey.Input{
				Message:	"TTL:",
				Help:		"Time to live in seconds for which a check result is valid",
				Default:	opts.TTL,
			},
		},
		{
			Name:	"subscriptions",
			Prompt: &survey.Input{
				Message:	"Subscriptions:",
				Default:	opts.Subscriptions,
			},
			Validate:	survey.Required,
		},
		{
			Name:	"handlers",
			Prompt: &survey.Input{
				Message:	"Handlers:",
				Default:	opts.Handlers,
			},
		},
		{
			Name:	"assets",
			Prompt: &survey.Input{
				Message:	"Runtime Assets:",
				Default:	opts.RuntimeAssets,
			},
		},
		{
			Name:	"publish",
			Prompt: &survey.Input{
				Message:	"Publish:",
				Default:	opts.Publish,
				Help:		"If check requests are published for the check. Value must be true or false.",
			},
			Validate: func(val interface{}) error {
				if value, ok := val.(string); ok {
					_, err := strconv.ParseBool(value)
					return err
				}
				return nil
			},
		},
		{
			Name:	"proxy-entity-name",
			Prompt: &survey.Input{
				Message:	"Check Proxy Entity Name:",
				Default:	opts.ProxyEntityName,
				Help:		"the check's proxy entity name, used to create a proxy entity for an external resource",
			},
		},
		{
			Name:	"stdin",
			Prompt: &survey.Input{
				Message:	"Check STDIN:",
				Default:	opts.Stdin,
				Help:		"If check accepts JSON event data to the check command's stdin. Defaults to false.",
			},
		},
		{
			Name:	"high-flap-threshold",
			Prompt: &survey.Input{
				Message:	"High Flap Threshold:",
				Default:	opts.HighFlapThreshold,
			},
		},
		{
			Name:	"low-flap-threshold",
			Prompt: &survey.Input{
				Message:	"Low Flap Threshold:",
				Default:	opts.LowFlapThreshold,
			},
		},
		{
			Name:	"output-metric-format",
			Prompt: &survey.Select{
				Message:	"Metric Format:",
				Options:	append([]string{"none"}, v2.OutputMetricFormats...),
				Default:	opts.OutputMetricFormat,
			},
			Validate: func(val interface{}) error {
				if value, _ := val.(string); value != "" && strings.TrimSpace(value) != "none" {
					if err := v2.ValidateOutputMetricFormat(value); err != nil {
						return err
					}
				}
				return nil
			},
		},
		{
			Name:	"output-metric-handlers",
			Prompt: &survey.Input{
				Message:	"Metric Handlers:",
				Default:	opts.OutputMetricHandlers,
			},
		},
		{
			Name:	"round-robin",
			Prompt: &survey.Input{
				Message:	"Round Robin",
				Default:	opts.RoundRobin,
				Help:		"if true, schedule this check in a round-robin fashion",
			},
			Validate: func(val interface{}) error {
				if value, ok := val.(string); ok {
					_, err := strconv.ParseBool(value)
					return err
				}
				return nil
			},
		},
	}...)

	return survey.Ask(qs, opts)
}

func (opts *checkOpts) Copy(check *v2.CheckConfig) {
	interval, _ := strconv.ParseUint(opts.Interval, 10, 32)
	stdin, _ := strconv.ParseBool(opts.Stdin)
	timeout, _ := strconv.ParseUint(opts.Timeout, 10, 32)
	ttl, _ := strconv.ParseInt(opts.TTL, 10, 64)
	highFlap, _ := strconv.ParseUint(opts.HighFlapThreshold, 10, 32)
	lowFlap, _ := strconv.ParseUint(opts.LowFlapThreshold, 10, 32)

	check.Name = opts.Name
	check.Namespace = opts.Namespace
	check.Interval = uint32(interval)
	check.Command = opts.Command
	check.Cron = opts.Cron
	check.Subscriptions = helpers.SafeSplitCSV(opts.Subscriptions)
	check.Handlers = helpers.SafeSplitCSV(opts.Handlers)
	check.RuntimeAssets = helpers.SafeSplitCSV(opts.RuntimeAssets)
	check.Publish, _ = strconv.ParseBool(opts.Publish)
	check.ProxyEntityName = opts.ProxyEntityName
	check.Stdin = stdin
	check.Timeout = uint32(timeout)
	check.Ttl = int64(ttl)
	check.HighFlapThreshold = uint32(highFlap)
	check.LowFlapThreshold = uint32(lowFlap)
	check.OutputMetricFormat = opts.OutputMetricFormat
	if check.OutputMetricFormat == "none" {
		check.OutputMetricFormat = ""
	}
	check.OutputMetricHandlers = helpers.SafeSplitCSV(opts.OutputMetricHandlers)
	check.RoundRobin, _ = strconv.ParseBool(opts.RoundRobin)
}

package hook

import (
	"fmt"
	"strconv"

	"github.com/AlecAivazis/survey"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/pflag"
)

const (
	timeoutDefault = "60"
)

const (
	stdinDefault = "false"
)

type hookOpts struct {
	Name    string `survey:"name"`
	Command string `survey:"command"`
	Timeout string `survey:"timeout"`
	Stdin   string `survey:"stdin"`
	Env     string
	Org     string
}

func newHookOpts() *hookOpts {
	opts := hookOpts{}
	opts.Timeout = timeoutDefault
	opts.Stdin = stdinDefault
	return &opts
}

func (opts *hookOpts) withHook(hook *types.HookConfig) {
	opts.Name = hook.Name
	opts.Org = hook.Organization
	opts.Env = hook.Environment
	opts.Command = hook.Command
	opts.Timeout = strconv.Itoa(int(hook.Timeout))
	opts.Stdin = strconv.FormatBool(hook.Stdin)
}

func (opts *hookOpts) withFlags(flags *pflag.FlagSet) {
	opts.Command, _ = flags.GetString("command")
	opts.Timeout, _ = flags.GetString("timeout")
	stdinBool, _ := flags.GetBool("stdin")
	opts.Stdin = strconv.FormatBool(stdinBool)

	if org, _ := flags.GetString("organization"); org != "" {
		opts.Org = org
	}
	if env, _ := flags.GetString("environment"); env != "" {
		opts.Env = env
	}
}

func (opts *hookOpts) administerQuestionnaire(editing bool) error {
	var qs = []*survey.Question{}

	if !editing {
		qs = append(qs, []*survey.Question{
			{
				Name: "name",
				Prompt: &survey.Input{
					Message: "Hook Name:",
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
			Name: "timeout",
			Prompt: &survey.Input{
				Message: "Timeout:",
				Default: opts.Timeout,
			},
		},
		{
			Name: "stdin",
			Prompt: &survey.Input{
				Message: "Stdin:",
				Help:    "If stdin is enabled for the hook. Value must be true or false.",
				Default: opts.Stdin,
			},
			Validate: func(val interface{}) error {
				if str := val.(string); str != "false" && str != "true" {
					return fmt.Errorf("Please enter either true or false")
				}
				return nil
			},
		},
	}...)

	return survey.Ask(qs, opts)
}

func (opts *hookOpts) Copy(hook *types.HookConfig) {
	timeout, _ := strconv.ParseUint(opts.Timeout, 10, 32)
	stdin, _ := strconv.ParseBool(opts.Stdin)

	hook.Name = opts.Name
	hook.Environment = opts.Env
	hook.Organization = opts.Org
	hook.Timeout = uint32(timeout)
	hook.Command = opts.Command
	hook.Stdin = stdin
}

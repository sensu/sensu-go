package hook

import (
	"fmt"
	"strconv"

	"github.com/AlecAivazis/survey/v2"
	"github.com/sensu/sensu-go/cli/commands/helpers"
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
	Name      string `survey:"name"`
	Command   string `survey:"command"`
	Timeout   string `survey:"timeout"`
	Stdin     string `survey:"stdin"`
	Env       string
	Namespace string
}

func newHookOpts() *hookOpts {
	opts := hookOpts{}
	opts.Timeout = timeoutDefault
	opts.Stdin = stdinDefault
	return &opts
}

func (opts *hookOpts) withHook(hook *types.HookConfig) {
	opts.Name = hook.Name
	opts.Namespace = hook.Namespace
	opts.Command = hook.Command
	opts.Timeout = strconv.Itoa(int(hook.Timeout))
	opts.Stdin = strconv.FormatBool(hook.Stdin)
}

func (opts *hookOpts) withFlags(flags *pflag.FlagSet) {
	opts.Command, _ = flags.GetString("command")
	opts.Timeout, _ = flags.GetString("timeout")
	stdinBool, _ := flags.GetBool("stdin")
	opts.Stdin = strconv.FormatBool(stdinBool)

	if namespace := helpers.GetChangedStringValueViper("namespace", flags); namespace != "" {
		opts.Namespace = namespace
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
				Name: "namespace",
				Prompt: &survey.Input{
					Message: "Namespace:",
					Default: opts.Namespace,
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
				if str, ok := val.(string); ok && str != "false" && str != "true" {
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
	hook.Namespace = opts.Namespace
	hook.Timeout = uint32(timeout)
	hook.Command = opts.Command
	hook.Stdin = stdin
}

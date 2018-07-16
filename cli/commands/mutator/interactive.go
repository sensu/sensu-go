package mutator

import (
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/pflag"
)

type mutatorOpts struct {
	Name    string `survey:"name"`
	Command string `survey:"command"`
	Timeout string `survey:"timeout"`
	EnvVars string `survey:"env-vars"`
	Env     string
	Org     string
}

func newMutatorOpts() *mutatorOpts {
	opts := mutatorOpts{}
	return &opts
}

func (opts *mutatorOpts) withMutator(mutator *types.Mutator) {
	opts.Name = mutator.Name
	opts.Env = mutator.Environment
	opts.Org = mutator.Organization

	opts.Command = mutator.Command
	opts.Timeout = strconv.FormatUint(uint64(mutator.Timeout), 10)
	opts.EnvVars = strings.Join(mutator.EnvVars, ",")
}

func (opts *mutatorOpts) withFlags(flags *pflag.FlagSet) {
	opts.Command, _ = flags.GetString("command")
	opts.Timeout, _ = flags.GetString("timeout")
	opts.EnvVars, _ = flags.GetString("env-vars")

	if org := helpers.GetChangedStringValueFlag("organization", flags); org != "" {
		opts.Org = org
	}
	if env := helpers.GetChangedStringValueFlag("environment", flags); env != "" {
		opts.Env = env
	}
}

func (opts *mutatorOpts) administerQuestionnaire(editing bool) error {
	var qs []*survey.Question
	if !editing {
		qs = append(qs, []*survey.Question{
			{
				Name: "name",
				Prompt: &survey.Input{
					Message: "Mutator Name:",
					Default: opts.Name},
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
		{Name: "command",
			Prompt: &survey.Input{
				Message: "Command:",
				Default: opts.Command,
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
			Name: "env-vars",
			Prompt: &survey.Input{
				Message: "Environment variables:",
				Help:    "A list of comma-separated key=value pairs of environment variables.",
				Default: opts.EnvVars,
			},
		},
	}...)

	return survey.Ask(qs, opts)
}

func (opts *mutatorOpts) Copy(mutator *types.Mutator) {
	mutator.Name = opts.Name
	mutator.Environment = opts.Env
	mutator.Organization = opts.Org

	mutator.Command = opts.Command
	mutator.EnvVars = helpers.SafeSplitCSV(opts.EnvVars)

	if len(opts.Timeout) > 0 {
		t, _ := strconv.ParseUint(opts.Timeout, 10, 32)
		mutator.Timeout = uint32(t)
	} else {
		mutator.Timeout = 0
	}
}

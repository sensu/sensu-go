package mutator

import (
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	v2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/spf13/pflag"
)

type mutatorOpts struct {
	Type		string	`survey:"type"`
	Name		string	`survey:"name"`
	Command		string	`survey:"command"`
	Eval		string	`survey:"eval"`
	Timeout		string	`survey:"timeout"`
	EnvVars		string	`survey:"env-vars"`
	Namespace	string	`survey:"namespace"`
	RuntimeAssets	string	`survey:"assets"`
}

func newMutatorOpts() *mutatorOpts {
	opts := mutatorOpts{}
	return &opts
}

func (opts *mutatorOpts) withMutator(mutator *v2.Mutator) {
	opts.Name = mutator.Name
	opts.Namespace = mutator.Namespace
	opts.Type = mutator.Type
	opts.Eval = mutator.Eval
	opts.Command = mutator.Command
	opts.Timeout = strconv.FormatUint(uint64(mutator.Timeout), 10)
	opts.EnvVars = strings.Join(mutator.EnvVars, ",")
	opts.RuntimeAssets = strings.Join(mutator.RuntimeAssets, ",")
}

func (opts *mutatorOpts) withFlags(flags *pflag.FlagSet) {
	opts.Command, _ = flags.GetString("command")
	opts.Timeout, _ = flags.GetString("timeout")
	opts.EnvVars, _ = flags.GetString("env-vars")
	opts.RuntimeAssets, _ = flags.GetString("runtime-assets")
	opts.Type, _ = flags.GetString("type")
	opts.Eval, _ = flags.GetString("eval")

	if namespace := helpers.GetChangedStringValueViper("namespace", flags); namespace != "" {
		opts.Namespace = namespace
	}
}

func (opts *mutatorOpts) administerQuestionnaire(editing bool) error {
	var qs []*survey.Question
	if !editing {
		qs = append(qs, []*survey.Question{
			{
				Name:	"name",
				Prompt: &survey.Input{
					Message:	"Mutator Name:",
					Default:	opts.Name},
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
			Name:	"type",
			Prompt: &survey.Input{
				Message:	"Type:",
				Default:	"pipe",
			},
		},
		{
			Name:	"command",
			Prompt: &survey.Input{
				Message:	"Command:",
				Default:	opts.Command,
			},
		},
		{
			Name:	"eval",
			Prompt: &survey.Input{
				Message: "Eval:",
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
			Name:	"env-vars",
			Prompt: &survey.Input{
				Message:	"Environment variables:",
				Help:		"A list of comma-separated key=value pairs of environment variables.",
				Default:	opts.EnvVars,
			},
		},
		{
			Name:	"assets",
			Prompt: &survey.Input{
				Message:	"Runtime Assets:",
				Help:		"A list of comma-separated list of assets to use when executing the mutator",
				Default:	opts.RuntimeAssets,
			},
		},
	}...)

	return survey.Ask(qs, opts)
}

func (opts *mutatorOpts) Copy(mutator *v2.Mutator) {
	mutator.Name = opts.Name
	mutator.Namespace = opts.Namespace
	mutator.Type = opts.Type
	mutator.Eval = opts.Eval
	mutator.Command = opts.Command
	mutator.EnvVars = helpers.SafeSplitCSV(opts.EnvVars)

	if len(opts.Timeout) > 0 {
		t, _ := strconv.ParseUint(opts.Timeout, 10, 32)
		mutator.Timeout = uint32(t)
	} else {
		mutator.Timeout = 0
	}

	assets := helpers.SafeSplitCSV(opts.RuntimeAssets)
	mutator.RuntimeAssets = make([]string, len(assets))
	for i, h := range assets {
		mutator.RuntimeAssets[i] = strings.TrimSpace(h)
	}
}

package filter

import (
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/pflag"
)

type filterOpts struct {
	Action        string `survey:"action"`
	Name          string `survey:"name"`
	Namespace     string `survey:"namespace"`
	Expressions   string `survey:"expressions"`
	RuntimeAssets string `survey:"runtimeAssets"`
}

func newFilterOpts() *filterOpts {
	return &filterOpts{}
}

func (opts *filterOpts) administerQuestionnaire(editing bool, askOpts ...survey.AskOpt) error {
	var qs = []*survey.Question{}

	if !editing {
		qs = append(qs, []*survey.Question{
			{
				Name: "name",
				Prompt: &survey.Input{
					Message: "Filter Name:",
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
			Name: "action",
			Prompt: &survey.Select{
				Message: "Action:",
				Options: types.EventFilterAllActions,
				Default: types.EventFilterAllActions[0],
			},
			Validate: survey.Required,
		},
		{
			Name: "expressions",
			Prompt: &survey.Input{
				Message: "Expressions (comma separated list):",
				Default: opts.Expressions,
			},
			Validate: survey.Required,
		},
		{
			Name: "runtimeAssets",
			Prompt: &survey.Input{
				Message: "Runtime Assets:",
				Default: "",
			},
		},
	}...)

	return survey.Ask(qs, opts, askOpts...)
}

func (opts *filterOpts) Copy(filter *types.EventFilter) {
	filter.Action = opts.Action
	filter.Name = opts.Name
	filter.Namespace = opts.Namespace
	filter.Expressions = helpers.SafeSplitCSV(opts.Expressions)
	filter.RuntimeAssets = helpers.SafeSplitCSV(opts.RuntimeAssets)
}

func (opts *filterOpts) withFilter(filter *types.EventFilter) {
	opts.Name = filter.Name
	opts.Namespace = filter.Namespace
	opts.Action = filter.Action
	opts.Expressions = strings.Join(filter.Expressions, ",")
	opts.RuntimeAssets = strings.Join(filter.RuntimeAssets, ",")
}

func (opts *filterOpts) withFlags(flags *pflag.FlagSet) {
	opts.Action, _ = flags.GetString("action")
	opts.Expressions, _ = flags.GetString("expressions")

	if namespace := helpers.GetChangedStringValueViper("namespace", flags); namespace != "" {
		opts.Namespace = namespace
	}
}

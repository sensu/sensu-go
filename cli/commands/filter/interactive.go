package filter

import (
	"strings"

	"github.com/AlecAivazis/survey"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/pflag"
)

type filterOpts struct {
	Action     string `survey:"action"`
	Name       string `survey:"name"`
	Namespace  string
	Statements string `survey:"statements"`
}

func newFilterOpts() *filterOpts {
	return &filterOpts{}
}

func (opts *filterOpts) administerQuestionnaire(editing bool) error {
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
			Name: "statements",
			Prompt: &survey.Input{
				Message: "Statements (comma separated list):",
				Default: opts.Statements,
			},
			Validate: survey.Required,
		},
	}...)

	return survey.Ask(qs, opts)
}

func (opts *filterOpts) Copy(filter *types.EventFilter) {
	filter.Action = opts.Action
	filter.Name = opts.Name
	filter.Namespace = opts.Namespace
	filter.Statements = helpers.SafeSplitCSV(opts.Statements)
}

func (opts *filterOpts) withFilter(filter *types.EventFilter) {
	opts.Name = filter.Name
	opts.Namespace = filter.Namespace
	opts.Action = filter.Action
	opts.Statements = strings.Join(filter.Statements, ",")
}

func (opts *filterOpts) withFlags(flags *pflag.FlagSet) {
	opts.Action, _ = flags.GetString("action")
	opts.Statements, _ = flags.GetString("statements")

	if namespace := helpers.GetChangedStringValueFlag("namespace", flags); namespace != "" {
		opts.Namespace = namespace
	}
}

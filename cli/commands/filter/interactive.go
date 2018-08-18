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
	Env        string
	Name       string `survey:"name"`
	Org        string
	Statements string `survey:"statements"`
	Type       string `survey:"type"`
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
		{
			Name: "type",
			Prompt: &survey.Input{
				Message: "Type (js or govaluate)",
				Default: opts.Type,
			},
			Validate: survey.Required,
		},
	}...)

	return survey.Ask(qs, opts)
}

func (opts *filterOpts) Copy(filter *types.EventFilter) {
	filter.Action = opts.Action
	filter.Environment = opts.Env
	filter.Name = opts.Name
	filter.Organization = opts.Org
	filter.Statements = helpers.SafeSplitCSV(opts.Statements)
	filter.Type = opts.Type
}

func (opts *filterOpts) withFilter(filter *types.EventFilter) {
	opts.Name = filter.Name
	opts.Org = filter.Organization
	opts.Env = filter.Environment
	opts.Action = filter.Action
	opts.Statements = strings.Join(filter.Statements, ",")
	opts.Type = filter.Type
}

func (opts *filterOpts) withFlags(flags *pflag.FlagSet) {
	opts.Action, _ = flags.GetString("action")
	opts.Statements, _ = flags.GetString("statements")
	opts.Type, _ = flags.GetString("type")

	if org := helpers.GetChangedStringValueFlag("organization", flags); org != "" {
		opts.Org = org
	}
	if env := helpers.GetChangedStringValueFlag("environment", flags); env != "" {
		opts.Env = env
	}
}

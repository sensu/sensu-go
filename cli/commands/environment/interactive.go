package environment

import (
	"github.com/AlecAivazis/survey"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/pflag"
)

type envOpts struct {
	Description string `survey:"description"`
	Name        string `survey:"name"`
	Org         string `survey:"organization"`
}

func (opts *envOpts) withEnv(env *types.Environment) {
	opts.Name = env.Name
	opts.Description = env.Description
}

func (opts *envOpts) withFlags(flags *pflag.FlagSet) {
	opts.Description, _ = flags.GetString("description")

	if org, _ := flags.GetString("organization"); org != "" {
		opts.Org = org
	}
}

func (opts *envOpts) administerQuestionnaire(editing bool) error {
	var qs []*survey.Question

	if !editing {
		qs = append(qs, []*survey.Question{
			{
				Name: "name",
				Prompt: &survey.Input{
					Message: "Name:",
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
		}...)
	}

	qs = append(qs, []*survey.Question{
		{
			Name: "description",
			Prompt: &survey.Input{
				Message: "Description:",
				Default: opts.Description,
			},
		},
	}...)

	return survey.Ask(qs, opts)
}

func (opts *envOpts) Copy(env *types.Environment) {
	env.Description = opts.Description
	env.Name = opts.Name
}

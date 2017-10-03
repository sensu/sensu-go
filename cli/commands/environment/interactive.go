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
	opts.Description = env.Description
	opts.Name = env.Name
}

func (opts *envOpts) withFlags(flags *pflag.FlagSet) {
	opts.Description, _ = flags.GetString("description")
	opts.Name, _ = flags.GetString("name")
	opts.Org, _ = flags.GetString("org")
}

func (opts *envOpts) administerQuestionnaire(editing bool) {
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

	survey.Ask(qs, opts)
}

func (opts *envOpts) Copy(env *types.Environment) {
	env.Description = opts.Description
	env.Name = opts.Name
}

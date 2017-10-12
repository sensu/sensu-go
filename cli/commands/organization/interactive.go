package organization

import (
	"github.com/AlecAivazis/survey"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/pflag"
)

type orgOpts struct {
	Description string `survey:"description"`
	Name        string `survey:"name"`
}

func newOrgOpts() *orgOpts {
	opts := orgOpts{}
	return &opts
}

func (opts *orgOpts) withOrg(org *types.Organization) {
	opts.Description = org.Description
	opts.Name = org.Name
}

func (opts *orgOpts) withFlags(flags *pflag.FlagSet) {
	opts.Description, _ = flags.GetString("description")
}

func (opts *orgOpts) administerQuestionnaire(editing bool) {
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
		}...)
	}

	qs = append(qs, []*survey.Question{
		{
			Name: "description",
			Prompt: &survey.Input{
				"Description:",
				opts.Description,
			},
		},
	}...)

	survey.Ask(qs, opts)
}

func (opts *orgOpts) Copy(org *types.Organization) {
	org.Description = opts.Description
	org.Name = opts.Name
}

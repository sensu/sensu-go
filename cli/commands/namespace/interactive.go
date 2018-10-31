package namespace

import (
	"github.com/AlecAivazis/survey"
	"github.com/sensu/sensu-go/types"
)

type namespaceOpts struct {
	Description string `survey:"description"`
	Name        string `survey:"name"`
}

func newNamespaceOpts() *namespaceOpts {
	opts := namespaceOpts{}
	return &opts
}

func (opts *namespaceOpts) administerQuestionnaire(editing bool) error {
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

	return survey.Ask(qs, opts)
}

func (opts *namespaceOpts) Copy(namespace *types.Namespace) {
	namespace.Name = opts.Name
}

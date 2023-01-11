package namespace

import (
	"github.com/AlecAivazis/survey/v2"
	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
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

func (opts *namespaceOpts) Copy(namespace *corev3.Namespace) {
	if namespace.Metadata == nil {
		namespace.Metadata = &corev2.ObjectMeta{
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
		}
	}
	namespace.Metadata.Name = opts.Name
}

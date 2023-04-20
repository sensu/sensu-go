package entity

import (
	"strings"

	"github.com/AlecAivazis/survey/v2"
	v2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/spf13/pflag"
)

type entityOpts struct {
	Name		string	`survey:"name"`
	EntityClass	string	`survey:"entity-class"`
	Subscriptions	string	`survey:"subscriptions"`
	Namespace	string
}

func newEntityOpts() *entityOpts {
	return &entityOpts{}
}

func (opts *entityOpts) withFlags(flags *pflag.FlagSet) {
	opts.EntityClass, _ = flags.GetString("entity-class")
	opts.Subscriptions, _ = flags.GetString("subscriptions")

	if namespace := helpers.GetChangedStringValueViper("namespace", flags); namespace != "" {
		opts.Namespace = namespace
	}
}

func (opts *entityOpts) administerQuestionnaire(editing bool) error {
	var qs = []*survey.Question{}

	if !editing {
		qs = append(qs, []*survey.Question{
			{
				Name:	"name",
				Prompt: &survey.Input{
					Message:	"Entity Name:",
					Default:	opts.Name,
				},
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
			Name:	"entity-class",
			Prompt: &survey.Input{
				Message:	"Entity Class:",
				Default:	opts.EntityClass,
				Help:		"entity class, either proxy or agent",
			},
			Validate:	survey.Required,
		},
		{
			Name:	"subscriptions",
			Prompt: &survey.Input{
				Message:	"Subscriptions:",
				Default:	opts.Subscriptions,
				Help:		"comma separated list of subscriptions",
			},
		},
	}...)

	return survey.Ask(qs, opts)
}

func (opts *entityOpts) copy(entity *v2.Entity) {
	entity.Name = opts.Name
	entity.EntityClass = opts.EntityClass
	entity.Subscriptions = helpers.SafeSplitCSV(opts.Subscriptions)
	entity.Namespace = opts.Namespace
}

func (opts *entityOpts) withEntity(entity *v2.Entity) {
	opts.Name = entity.Name
	opts.EntityClass = entity.EntityClass
	opts.Subscriptions = strings.Join(entity.Subscriptions, ",")
	opts.Namespace = entity.Namespace
}

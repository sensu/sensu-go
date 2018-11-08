package entity

import (
	"strings"

	"github.com/AlecAivazis/survey"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/pflag"
)

type entityOpts struct {
	Name          string `survey:"name"`
	EntityClass   string `survey:"entity-class"`
	Subscriptions string `survey:"subscriptions"`
	Namespace     string
}

func newEntityOpts() *entityOpts {
	return &entityOpts{}
}

func (opts *entityOpts) withFlags(flags *pflag.FlagSet) {
	opts.EntityClass, _ = flags.GetString("entity-class")
	opts.Subscriptions, _ = flags.GetString("subscriptions")

	if namespace := helpers.GetChangedStringValueFlag("namespace", flags); namespace != "" {
		opts.Namespace = namespace
	}
}

func (opts *entityOpts) administerQuestionnaire() error {
	qs := []*survey.Question{
		{
			Name: "class",
			Prompt: &survey.Input{
				Message: "Entity Class:",
				Default: opts.EntityClass,
				Help:    "entity class, either proxy or agent",
			},
			Validate: survey.Required,
		},
		{
			Name: "subscriptions",
			Prompt: &survey.Input{
				Message: "Subscriptions:",
				Default: opts.Subscriptions,
				Help:    "comma separated list of subscriptions",
			},
		},
		{
			Name: "namespace",
			Prompt: &survey.Input{
				Message: "Namespace:",
				Default: opts.Namespace,
			},
			Validate: survey.Required,
		},
	}

	return survey.Ask(qs, opts)
}

func (opts *entityOpts) copy(entity *types.Entity) {
	entity.Name = opts.Name
	entity.EntityClass = opts.EntityClass
	entity.Subscriptions = helpers.SafeSplitCSV(opts.Subscriptions)
	entity.Namespace = opts.Namespace
}

func (opts *entityOpts) withEntity(entity *types.Entity) {
	opts.Name = entity.Name
	opts.EntityClass = entity.EntityClass
	opts.Subscriptions = strings.Join(entity.Subscriptions, ",")
	opts.Namespace = entity.Namespace
}

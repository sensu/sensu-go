package entity

import (
	"strings"

	"github.com/AlecAivazis/survey"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/pflag"
)

type entityOpts struct {
	ID            string `survey:"id"`
	Class         string `survey:"class"`
	Subscriptions string `survey:"subscriptions"`
	Org           string
	Env           string
}

func newEntityOpts() *entityOpts {
	return &entityOpts{}
}

func (opts *entityOpts) withFlags(flags *pflag.FlagSet) {
	opts.Class, _ = flags.GetString("class")
	opts.Subscriptions, _ = flags.GetString("subscriptions")
}

func (opts *entityOpts) administerQuestionnaire() error {
	qs := []*survey.Question{
		{
			Name: "subscriptions",
			Prompt: &survey.Input{
				Message: "Subscriptions:",
				Default: opts.Subscriptions,
				Help:    "comma separated list of subscriptions",
			},
			Validate: survey.Required,
		},
	}

	return survey.Ask(qs, opts)
}

func (opts *entityOpts) copy(entity *types.Entity) {
	entity.ID = opts.ID
	entity.Class = opts.Class
	entity.Subscriptions = helpers.SafeSplitCSV(opts.Subscriptions)
	entity.Environment = opts.Env
	entity.Organization = opts.Org
}

func (opts *entityOpts) withEntity(entity *types.Entity) {
	opts.ID = entity.ID
	opts.Class = entity.Class
	opts.Subscriptions = strings.Join(entity.Subscriptions, ",")
	opts.Env = entity.Environment
	opts.Org = entity.Organization
}

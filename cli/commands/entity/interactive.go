package entity

import (
	"strings"

	"github.com/AlecAivazis/survey"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/types"
)

type opts struct {
	Subscriptions string `survey:"subscriptions"`
}

func newOpts() *opts {
	return &opts{}
}

func (opts *opts) administerQuestionnaire() error {
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

func (opts *opts) copy(entity *types.Entity) {
	entity.Subscriptions = helpers.SafeSplitCSV(opts.Subscriptions)
}

func (opts *opts) withEntity(entity *types.Entity) {
	opts.Subscriptions = strings.Join(entity.Subscriptions, ",")
}

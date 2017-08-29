package user

import (
	"strings"

	"github.com/AlecAivazis/survey"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/pflag"
)

type userOpts struct {
	Username string `survey:"username"`
	Password string `survey:"password"`
	Roles    string `survey:"roles"`
}

func newUserOpts() *userOpts {
	opts := userOpts{}
	return &opts
}

func (opts *userOpts) withUser(user *types.User) {
	opts.Username = user.Username
	opts.Password = user.Password
	opts.Roles = strings.Join(user.Roles, ",")
}

func (opts *userOpts) withFlags(flags *pflag.FlagSet) {
	opts.Username, _ = flags.GetString("username")
	opts.Password, _ = flags.GetString("password")
	opts.Roles, _ = flags.GetString("roles")
}

func (opts *userOpts) administerQuestionnaire() {
	var qs = []*survey.Question{
		{
			Name: "username",
			Prompt: &survey.Input{
				"Username:",
				opts.Username,
			},
			Validate: survey.Required,
		},
		{
			Name: "password",
			Prompt: &survey.Password{
				Message: "Password:",
			},
			Validate: survey.Required,
		},
		{
			Name: "roles",
			Prompt: &survey.Input{
				"Roles:",
				opts.Roles,
			},
		},
	}

	survey.Ask(qs, opts)
}

func (opts *userOpts) Copy(user *types.User) {
	user.Username = opts.Username
	user.Password = opts.Password
	user.Roles = helpers.SafeSplitCSV(opts.Roles)
}

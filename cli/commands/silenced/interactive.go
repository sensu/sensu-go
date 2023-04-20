package silenced

import (
	"fmt"
	"io"
	"strconv"

	"github.com/AlecAivazis/survey/v2"
	v2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/commands/timeutil"
	"github.com/spf13/pflag"
)

const (
	expireDefault	= "-1"
	beginDefault	= "0"
)

type silencedOpts struct {
	Check		string	`survey:"check"`
	Subscription	string	`survey:"subscription"`
	Expire		string	`survey:"expire"`
	ExpireOnResolve	bool	`survey:"expire_on_resolve"`
	Creator		string
	Reason		string	`survey:"reason"`
	Env		string
	Namespace	string
	Begin		string	`survey:"begin"`
}

func newSilencedOpts() *silencedOpts {
	opts := silencedOpts{}
	opts.Expire = expireDefault
	opts.Begin = beginDefault
	return &opts
}

func (o *silencedOpts) Apply(s *v2.Silenced) (err error) {
	s.Subscription = o.Subscription
	s.Check = o.Check
	s.Creator = o.Creator
	s.Reason = o.Reason
	s.Namespace = o.Namespace
	s.ExpireOnResolve = o.ExpireOnResolve
	s.Expire, err = strconv.ParseInt(o.Expire, 10, 64)
	if err != nil {
		return err
	}
	s.Begin, err = timeutil.ConvertToUnix(o.Begin)
	return err
}

func (o *silencedOpts) withFlags(flags *pflag.FlagSet) {
	o.Expire, _ = flags.GetString("expire")
	o.ExpireOnResolve, _ = flags.GetBool("expire-on-resolve")
	o.Reason, _ = flags.GetString("reason")
	o.Subscription, _ = flags.GetString("subscription")
	o.Check, _ = flags.GetString("check")
	o.Begin, _ = flags.GetString("begin")

	if namespace := helpers.GetChangedStringValueViper("namespace", flags); namespace != "" {
		o.Namespace = namespace
	}
}

func (o *silencedOpts) administerQuestionnaire(editing bool) error {
	var qs []*survey.Question

	if !editing {
		qs = []*survey.Question{
			{
				Name:	"namespace",
				Prompt: &survey.Input{
					Message:	"Namespace:",
					Default:	o.Namespace,
				},
				Validate:	survey.Required,
			},
			{
				Name:	"subscription",
				Prompt: &survey.Input{
					Message:	"Subscription:",
					Default:	o.Subscription,
					Help:		"One of subscription or check is required.",
				},
			},
			{
				Name:	"check",
				Prompt: &survey.Input{
					Message:	"Check:",
					Default:	o.Check,
					Help:		"One of subscription or check is required.",
				},
			},
		}
	}
	qs = append(qs, []*survey.Question{
		{
			Name:	"begin",
			Prompt: &survey.Input{
				Message:	"Begin time:",
				Default:	"now",
				Help:		"Start silencing events at this time. Format: Jan 02 2006 3:04PM MST",
			},
			Validate: func(val interface{}) error {
				if value, ok := val.(string); ok {
					_, err := timeutil.ConvertToUnix(value)
					return err
				}
				return nil
			},
		},
		{
			Name:	"expire",
			Prompt: &survey.Input{
				Message:	"Expiry in Seconds:",
				Default:	o.Expire,
			},
		},
		{
			Name:	"expire_on_resolve",
			Prompt: &survey.Confirm{
				Message:	"Expire on Resolve:",
				Default:	o.ExpireOnResolve,
				Help:		"Clear the silenced entry on resolution if true.",
			},
		},
		{
			Name:	"reason",
			Prompt: &survey.Input{
				Message:	"Reason:",
				Default:	o.Reason,
			},
		}}...)

	if err := survey.Ask(qs, o); err != nil && err != io.EOF {
		return err
	}
	return nil
}

type silencedName struct {
	Subscription	string
	Check		string
}

func askName(help string) (string, error) {
	questions := []*survey.Question{
		{
			Name:	"Subscription",
			Prompt: &survey.Input{
				Message:	"Subscription:",
				Help:		help,
			},
		},
		{
			Name:	"Check",
			Prompt: &survey.Input{
				Message:	"Check:",
				Help:		help,
			},
		},
	}

	var name silencedName
	if err := survey.Ask(questions, &name); err != nil {
		return "", err
	}
	return v2.SilencedName(name.Subscription, name.Check)
}

func toOpts(s *v2.Silenced) *silencedOpts {
	var o silencedOpts
	o.Subscription = s.Subscription
	o.Check = s.Check
	o.Creator = s.Creator
	o.Reason = s.Reason
	o.Namespace = s.Namespace
	o.ExpireOnResolve = s.ExpireOnResolve
	o.Expire = fmt.Sprintf("%d", s.Expire)
	o.Begin = fmt.Sprintf("%d", s.Begin)
	return &o
}

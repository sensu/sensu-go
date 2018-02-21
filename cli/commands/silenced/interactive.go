package silenced

import (
	"fmt"
	"io"
	"strconv"

	"github.com/AlecAivazis/survey"
	"github.com/sensu/sensu-go/cli/commands/timeutil"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/pflag"
)

const (
	expireDefault = "-1"
	beginDefault  = "0"
)

type silencedOpts struct {
	Check           string `survey:"check"`
	Subscription    string `survey:"subscription"`
	Expire          string `survey:"expire"`
	ExpireOnResolve bool   `survey:"expire_on_resolve"`
	Creator         string
	Reason          string `survey:"reason"`
	Env             string
	Org             string
	Begin           string `survey:"begin"`
}

func newSilencedOpts() *silencedOpts {
	opts := silencedOpts{}
	opts.Expire = expireDefault
	opts.Begin = beginDefault
	return &opts
}

func (o *silencedOpts) Apply(s *types.Silenced) (err error) {
	s.Subscription = o.Subscription
	s.Check = o.Check
	s.Creator = o.Creator
	s.Reason = o.Reason
	s.Environment = o.Env
	s.Organization = o.Org
	s.ExpireOnResolve = o.ExpireOnResolve
	s.Expire, err = strconv.ParseInt(o.Expire, 10, 64)
	if err != nil {
		return err
	}
	s.Begin, err = timeutil.ConvertToUnixUTC(o.Begin)
	return err
}

func (o *silencedOpts) withFlags(flags *pflag.FlagSet) (err error) {
	o.Expire, err = flags.GetString("expire")
	if err != nil {
		return err
	}
	o.ExpireOnResolve, err = flags.GetBool("expire-on-resolve")
	if err != nil {
		return err
	}
	o.Reason, err = flags.GetString("reason")
	if err != nil {
		return err
	}
	o.Subscription, err = flags.GetString("subscription")
	if err != nil {
		return err
	}
	o.Check, err = flags.GetString("check")
	if err != nil {
		return err
	}
	o.Begin, err = flags.GetString("begin")
	return err
}

func (o *silencedOpts) administerQuestionnaire(editing bool) error {
	var qs []*survey.Question

	if !editing {
		qs = []*survey.Question{
			{
				Name: "org",
				Prompt: &survey.Input{
					Message: "Organization:",
					Default: o.Org,
				},
				Validate: survey.Required,
			},
			{
				Name: "env",
				Prompt: &survey.Input{
					Message: "Environment:",
					Default: o.Env,
				},
				Validate: survey.Required,
			},
			{
				Name: "subscription",
				Prompt: &survey.Input{
					Message: "Subscription:",
					Default: o.Subscription,
					Help:    "One of subscription or check is required.",
				},
			},
			{
				Name: "check",
				Prompt: &survey.Input{
					Message: "Check:",
					Default: o.Check,
					Help:    "One of subscription or check is required.",
				},
			},
		}
	}
	qs = append(qs, []*survey.Question{
		{
			Name: "begin",
			Prompt: &survey.Input{
				Message: "Begin time:",
				Default: o.Begin,
				Help:    "Start silencing events at this time. Format: Jan 02 2006 3:04PM MST",
			},
			Validate: func(val interface{}) error {
				_, err := timeutil.ConvertToUnixUTC(val.(string))
				return err
			},
		},
		{
			Name: "expire",
			Prompt: &survey.Input{
				Message: "Expiry in Seconds:",
				Default: o.Expire,
			},
		},
		{
			Name: "expire_on_resolve",
			Prompt: &survey.Confirm{
				Message: "Expire on Resolve:",
				Default: o.ExpireOnResolve,
				Help:    "Clear the silenced entry on resolution if true.",
			},
		},
		{
			Name: "reason",
			Prompt: &survey.Input{
				Message: "Reason:",
				Default: o.Reason,
			},
		}}...)

	if err := survey.Ask(qs, o); err != nil && err != io.EOF {
		return err
	}
	return nil
}

type silencedID struct {
	Subscription string
	Check        string
}

func askID(help string) (string, error) {
	questions := []*survey.Question{
		{
			Name: "Subscription",
			Prompt: &survey.Input{
				Message: "Subscription:",
				Help:    help,
			},
		},
		{
			Name: "Check",
			Prompt: &survey.Input{
				Message: "Check:",
				Help:    help,
			},
		},
	}

	var id silencedID
	if err := survey.Ask(questions, &id); err != nil {
		return "", err
	}
	return types.SilencedID(id.Subscription, id.Check)
}

func toOpts(s *types.Silenced) *silencedOpts {
	var o silencedOpts
	o.Subscription = s.Subscription
	o.Check = s.Check
	o.Creator = s.Creator
	o.Reason = s.Reason
	o.Env = s.Environment
	o.Org = s.Organization
	o.ExpireOnResolve = s.ExpireOnResolve
	o.Expire = fmt.Sprintf("%d", s.Expire)
	o.Begin = fmt.Sprintf("%d", s.Begin)
	return &o
}

package silenced

import (
	"fmt"
	"io"
	"strconv"

	"github.com/AlecAivazis/survey"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/pflag"
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
	return err
}

func (o *silencedOpts) withFlags(flags *pflag.FlagSet) (err error) {
	exp, err := flags.GetInt64("expire")
	if err != nil {
		return err
	}
	o.Expire = fmt.Sprintf("%d", exp)
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
			Name: "expire",
			Prompt: &survey.Input{
				Message: "Expiry in Seconds:",
				Default: o.Expire,
			},
			Validate: survey.Required,
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
	return &o
}

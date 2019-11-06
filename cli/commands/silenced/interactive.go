package silenced

import (
	"fmt"
	"io"
	"strconv"

	"github.com/AlecAivazis/survey"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/commands/timeutil"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/pflag"
)

const (
	expireDefault = "-1"
	beginDefault  = "0"
)

type silencedOpts struct {
	Name            string `survey:"name"`
	Check           string `survey:"check"`
	Subscription    string `survey:"subscription"`
	Expire          string `survey:"expire"`
	ExpireOnResolve bool   `survey:"expire_on_resolve"`
	Creator         string
	Reason          string `survey:"reason"`
	Env             string
	Namespace       string
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
	s.ObjectMeta.Name = o.Name
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
	o.Name, _ = flags.GetString("name")
	o.Check, _ = flags.GetString("check")
	o.Begin, _ = flags.GetString("begin")

	if namespace := helpers.GetChangedStringValueFlag("namespace", flags); namespace != "" {
		o.Namespace = namespace
	}
}

func (o *silencedOpts) administerQuestionnaire(editing bool) error {
	var qs []*survey.Question

	if !editing {
		qs = []*survey.Question{
			{
				Name: "namespace",
				Prompt: &survey.Input{
					Message: "Namespace:",
					Default: o.Namespace,
				},
				Validate: survey.Required,
			},
			{
				Name: "name",
				Prompt: &survey.Input{
					Message: "Name:",
					Default: o.Name,
					Help:    "Name to give the silence, must be unique. Backend will generate a unique name if omitted.",
				},
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
				Default: "now",
				Help:    "Start silencing events at this time. Format: Jan 02 2006 3:04PM MST",
			},
			Validate: func(val interface{}) error {
				_, err := timeutil.ConvertToUnix(val.(string))
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

func toOpts(s *types.Silenced) *silencedOpts {
	var o silencedOpts
	o.Name = s.ObjectMeta.Name
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

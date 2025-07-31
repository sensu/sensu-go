package logging

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/sensu/sensu-go/util/logging"
	"github.com/spf13/pflag"
	"io"
)

type loggingOpts struct {
	ModuleName string `survey:"module"`
	LogLevel   string `survey:"level"`
}

func newLoggingOpts() *loggingOpts {
	opts := loggingOpts{}
	return &opts
}

func (o *loggingOpts) Apply(s *logging.LogLevelRequest) (err error) {
	s.Level = o.LogLevel
	s.Module = o.ModuleName
	return nil
}

func (o *loggingOpts) withFlags(flags *pflag.FlagSet) {
	o.LogLevel, _ = flags.GetString("level")
	o.ModuleName, _ = flags.GetString("module")
}

// administerQuestionnaire will ask questions for setting log level for a module
func (o *loggingOpts) administerQuestionnaire(askOpts ...survey.AskOpt) error {
	var qs []*survey.Question
	qs = []*survey.Question{
		{
			Name: "module",
			Prompt: &survey.Input{
				Message: "Module Name:",
				Default: o.ModuleName,
			},
			Validate: survey.Required,
		},
		{
			Name: "level",
			Prompt: &survey.Select{
				Message: "Select Log level:",
				Options: []string{
					"warn",
					"info",
					"debug",
					"error",
					"fatal",
					"panic",
					"trace",
				},
				Default: "info",
			},
			Validate: survey.Required,
		},
	}

	if err := survey.Ask(qs, o, askOpts...); err != nil && err != io.EOF {
		return err
	}
	return nil
}

// administerQuestionnaireAllModule will ask questions for setting log level for all modules
func (o *loggingOpts) administerQuestionnaireAllModule(askOpts ...survey.AskOpt) error {
	var qs []*survey.Question
	qs = []*survey.Question{
		{
			Name: "level",
			Prompt: &survey.Select{
				Message: "Select Log level:",
				Options: []string{
					"warn",
					"info",
					"debug",
					"error",
					"fatal",
					"panic",
					"trace",
				},
				Default: "info",
			},
			Validate: survey.Required,
		},
	}

	if err := survey.Ask(qs, o, askOpts...); err != nil && err != io.EOF {
		return err
	}
	return nil
}

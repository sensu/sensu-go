package handler

import (
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/pflag"
)

type handlerOpts struct {
	Name       string `survey:"name"`
	Type       string `survey:"type"`
	Mutator    string `survey:"mutator"`
	Command    string `survey:"command"`
	Timeout    string `survey:"timeout"`
	Handlers   string `survey:"handler"`
	SocketHost string `survey:"socketHost"`
	SocketPort string `survey:"socketPort"`
	Env        string
	Org        string
}

const (
	typeDefault = "pipe"
)

func newHandlerOpts() *handlerOpts {
	opts := handlerOpts{}
	opts.Type = typeDefault
	return &opts
}

func (opts *handlerOpts) withHandler(handler *types.Handler) {
	opts.Name = handler.Name
	opts.Type = handler.Type
	opts.Mutator = handler.Mutator
	opts.Org = handler.Organization
	opts.Env = handler.Environment
	opts.Command = handler.Command
	opts.Timeout = strconv.Itoa(handler.Timeout)
	opts.Handlers = strings.Join(handler.Handlers, ",")
	opts.SocketHost = handler.Socket.Host
	opts.SocketPort = strconv.Itoa(handler.Socket.Port)
}

func (opts *handlerOpts) withFlags(flags *pflag.FlagSet) {
	opts.Type, _ = flags.GetString("type")
	opts.Mutator, _ = flags.GetString("mutator")
	opts.Command, _ = flags.GetString("command")
	opts.Timeout, _ = flags.GetString("timeout")
	opts.SocketHost, _ = flags.GetString("socket-host")
	opts.SocketPort, _ = flags.GetString("socket-port")
	opts.Handlers, _ = flags.GetString("handlers")
	opts.Org, _ = flags.GetString("organization")
	opts.Env, _ = flags.GetString("environment")
}

func (opts *handlerOpts) administerQuestionnaire(editing bool) {

	opts.queryForBaseParameters(editing)

	switch opts.Type {
	case "pipe":
		opts.queryForCommand()
	case "tcp":
		fallthrough
	case "udp":
		opts.queryForSocket()
	case "set":
		opts.queryForHandlers()
	}
}

func (opts *handlerOpts) queryForBaseParameters(editing bool) {
	var qs []*survey.Question

	if !editing {
		qs = append(qs, []*survey.Question{
			{
				Name:     "name",
				Prompt:   &survey.Input{"Handler Name:", ""},
				Validate: survey.Required,
			},
			{
				Name: "org",
				Prompt: &survey.Input{
					"Organization:",
					opts.Org,
				},
				Validate: survey.Required,
			},
			{
				Name: "env",
				Prompt: &survey.Input{
					"Environment:",
					opts.Env,
				},
				Validate: survey.Required,
			},
		}...)
	}

	qs = append(qs, []*survey.Question{
		{
			Name:   "mutator",
			Prompt: &survey.Input{"Mutator:", opts.Mutator},
		},
		{
			Name:   "timeout",
			Prompt: &survey.Input{"Timeout:", opts.Timeout},
		},
		{
			Name: "type",
			Prompt: &survey.Select{
				Message: "Type:",
				Options: []string{"pipe", "tcp", "udp", "set"},
				Default: opts.Type,
			},
			Validate: survey.Required,
		},
	}...)

	survey.Ask(qs, opts)
}

func (opts *handlerOpts) queryForCommand() {
	var qs = []*survey.Question{
		{
			Name:     "command",
			Prompt:   &survey.Input{"Command:", opts.Command},
			Validate: survey.Required,
		},
	}

	survey.Ask(qs, opts)
}

func (opts *handlerOpts) queryForSocket() {
	var qs = []*survey.Question{
		{
			Name:     "socketHost",
			Prompt:   &survey.Input{"Socket Host:", opts.SocketHost},
			Validate: survey.Required,
		},
		{
			Name:     "socketPort",
			Prompt:   &survey.Input{"Socket Port:", opts.SocketPort},
			Validate: survey.Required,
		},
	}

	survey.Ask(qs, opts)
}

func (opts *handlerOpts) queryForHandlers() {
	var qs = []*survey.Question{
		{
			Name:     "handlers",
			Prompt:   &survey.Input{"Handlers:", opts.Handlers},
			Validate: survey.Required,
		},
	}

	survey.Ask(qs, opts)
}

func (opts *handlerOpts) Copy(handler *types.Handler) {
	handler.Name = opts.Name
	handler.Environment = opts.Env
	handler.Organization = opts.Org
	handler.Type = strings.ToLower(opts.Type)
	handler.Mutator = opts.Mutator
	handler.Command = opts.Command

	if len(opts.Timeout) > 0 {
		t, _ := strconv.Atoi(opts.Timeout)
		handler.Timeout = t
	} else {
		handler.Timeout = 0
	}

	if len(opts.SocketHost) > 0 && len(opts.SocketPort) > 0 {
		p, _ := strconv.Atoi(opts.SocketPort)
		handler.Socket = types.HandlerSocket{
			Host: opts.SocketHost,
			Port: p,
		}
	}

	handlers := helpers.SafeSplitCSV(opts.Handlers)
	handler.Handlers = make([]string, len(handlers))
	for i, h := range handlers {
		handler.Handlers[i] = strings.TrimSpace(h)
	}
}

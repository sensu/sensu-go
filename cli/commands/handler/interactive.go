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

	if org, _ := flags.GetString("organization"); org != "" {
		opts.Org = org
	}
	if env, _ := flags.GetString("environment"); env != "" {
		opts.Env = env
	}
}

func (opts *handlerOpts) administerQuestionnaire(editing bool) error {

	if err := opts.queryForBaseParameters(editing); err != nil {
		return err
	}

	switch opts.Type {
	case "pipe":
		return opts.queryForCommand()
	case "tcp":
		fallthrough
	case "udp":
		return opts.queryForSocket()
	case "set":
		return opts.queryForHandlers()
	}

	return nil
}

func (opts *handlerOpts) queryForBaseParameters(editing bool) error {
	var qs []*survey.Question

	if !editing {
		qs = append(qs, []*survey.Question{
			{
				Name: "name",
				Prompt: &survey.Input{
					Message: "Handler Name:",
					Default: opts.Name},
				Validate: survey.Required,
			},
			{
				Name: "org",
				Prompt: &survey.Input{
					Message: "Organization:",
					Default: opts.Org,
				},
				Validate: survey.Required,
			},
			{
				Name: "env",
				Prompt: &survey.Input{
					Message: "Environment:",
					Default: opts.Env,
				},
				Validate: survey.Required,
			},
		}...)
	}

	qs = append(qs, []*survey.Question{
		{
			Name: "mutator",
			Prompt: &survey.Input{
				Message: "Mutator:",
				Default: opts.Mutator,
			},
		},
		{
			Name: "timeout",
			Prompt: &survey.Input{
				Message: "Timeout:",
				Default: opts.Timeout,
			},
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

	return survey.Ask(qs, opts)
}

func (opts *handlerOpts) queryForCommand() error {
	var qs = []*survey.Question{
		{
			Name: "command",
			Prompt: &survey.Input{
				Message: "Command:",
				Default: opts.Command,
			},
			Validate: survey.Required,
		},
	}

	return survey.Ask(qs, opts)
}

func (opts *handlerOpts) queryForSocket() error {
	var qs = []*survey.Question{
		{
			Name: "socketHost",
			Prompt: &survey.Input{
				Message: "Socket Host:",
				Default: opts.SocketHost,
			},
			Validate: survey.Required,
		},
		{
			Name: "socketPort",
			Prompt: &survey.Input{
				Message: "Socket Port:",
				Default: opts.SocketPort,
			},
			Validate: survey.Required,
		},
	}

	return survey.Ask(qs, opts)
}

func (opts *handlerOpts) queryForHandlers() error {
	var qs = []*survey.Question{
		{
			Name: "handlers",
			Prompt: &survey.Input{
				Message: "Handlers:", Default: opts.Handlers,
			},
			Validate: survey.Required,
		},
	}

	return survey.Ask(qs, opts)
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

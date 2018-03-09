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
	Filters    string `survey:"filters"`
	Handlers   string `survey:"handlers"`
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
	opts.Env = handler.Environment
	opts.Org = handler.Organization

	opts.Command = handler.Command
	opts.Filters = strings.Join(handler.Filters, ",")
	opts.Handlers = strings.Join(handler.Handlers, ",")
	opts.Mutator = handler.Mutator
	opts.Timeout = strconv.FormatUint(uint64(handler.Timeout), 10)
	opts.Type = handler.Type

	if handler.Socket != nil {
		opts.SocketHost = handler.Socket.Host
		opts.SocketPort = strconv.FormatUint(uint64(handler.Socket.Port), 10)
	}
}

func (opts *handlerOpts) withFlags(flags *pflag.FlagSet) {
	opts.Command, _ = flags.GetString("command")
	opts.Filters, _ = flags.GetString("filters")
	opts.Handlers, _ = flags.GetString("handlers")
	opts.Mutator, _ = flags.GetString("mutator")
	opts.SocketHost, _ = flags.GetString("socket-host")
	opts.SocketPort, _ = flags.GetString("socket-port")
	opts.Timeout, _ = flags.GetString("timeout")
	opts.Type, _ = flags.GetString("type")

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
	case types.HandlerPipeType:
		return opts.queryForCommand()
	case types.HandlerTCPType:
		fallthrough
	case types.HandlerUDPType:
		return opts.queryForSocket()
	case types.HandlerSetType:
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
			Name: "filters",
			Prompt: &survey.Input{
				Message: "Filters:",
				Default: opts.Filters,
				Help:    "comma separated list of filters to use when filtering events for the handler",
			},
		},
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

func (opts *handlerOpts) queryForHandlers() error {
	var qs = []*survey.Question{
		{
			Name: "handlers",
			Prompt: &survey.Input{
				Message: "Handlers:",
				Default: opts.Handlers,
				Help:    "comma separated list of handlers to call using the handler set",
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

func (opts *handlerOpts) Copy(handler *types.Handler) {
	handler.Name = opts.Name
	handler.Environment = opts.Env
	handler.Organization = opts.Org

	handler.Command = opts.Command
	handler.Mutator = opts.Mutator
	handler.Type = strings.ToLower(opts.Type)

	if len(opts.Timeout) > 0 {
		t, _ := strconv.ParseUint(opts.Timeout, 10, 32)
		handler.Timeout = uint32(t)
	} else {
		handler.Timeout = 0
	}

	if len(opts.SocketHost) > 0 && len(opts.SocketPort) > 0 {
		p, _ := strconv.ParseUint(opts.SocketPort, 10, 32)
		handler.Socket = &types.HandlerSocket{
			Host: opts.SocketHost,
			Port: uint32(p),
		}
	}

	filters := helpers.SafeSplitCSV(opts.Filters)
	handler.Filters = make([]string, len(filters))
	for i, f := range filters {
		handler.Filters[i] = strings.TrimSpace(f)
	}

	handlers := helpers.SafeSplitCSV(opts.Handlers)
	handler.Handlers = make([]string, len(handlers))
	for i, h := range handlers {
		handler.Handlers[i] = strings.TrimSpace(h)
	}
}

package handler

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
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
}

// CreateCommand adds command that allows the user to create new handlers
func CreateCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create NAME",
		Short:        "create new handlers",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			flags := cmd.Flags()
			isInteractive := flags.NFlag() == 0
			opts := &handlerOpts{}

			if len(args) > 0 {
				opts.Name = args[0]
			}

			if isInteractive {
				opts.administerQuestionnaire()
			} else {
				opts.withFlags(flags)
			}

			handler := opts.toHandler()
			if err := handler.Validate(); err != nil {
				if !isInteractive {
					cmd.Help()
				}
				return err
			}

			err := cli.Client.CreateHandler(handler)
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "OK")
			return nil
		},
	}

	cmd.Flags().StringP("type", "t", "pipe", "type of handler (pipe, tcp, udp, or set)")
	cmd.Flags().StringP("mutator", "m", "", "Sensu event mutator (name) to use to mutate event data for the handler")
	cmd.Flags().StringP("command", "c", "", "command to be executed. The event data is passed to the process via STDIN")
	cmd.Flags().StringP("timeout", "i", "", "execution duration timeout in seconds (hard stop)")
	cmd.Flags().String("socket-host", "", "host of handler socket")
	cmd.Flags().String("socket-port", "", "port of handler socket")
	cmd.Flags().StringP("handlers", "", "", "comma separated list of handlers to call")

	return cmd
}

func (opts *handlerOpts) withFlags(flags *pflag.FlagSet) {
	opts.Type, _ = flags.GetString("type")
	opts.Mutator, _ = flags.GetString("mutator")
	opts.Command, _ = flags.GetString("command")
	opts.Timeout, _ = flags.GetString("timeout")
	opts.SocketHost, _ = flags.GetString("socket-host")
	opts.SocketPort, _ = flags.GetString("socket-port")
	opts.Handlers, _ = flags.GetString("handlers")
}

func (opts *handlerOpts) administerQuestionnaire() {
	opts.queryForBaseParameters()

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

func (opts *handlerOpts) queryForBaseParameters() {
	var qs = []*survey.Question{
		{
			Name:     "name",
			Prompt:   &survey.Input{"Handler Name:", ""},
			Validate: survey.Required,
		},
		{
			Name:   "mutator",
			Prompt: &survey.Input{"Mutator:", ""},
		},
		{
			Name:   "timeout",
			Prompt: &survey.Input{"Timeout:", ""},
		},
		{
			Name: "type",
			Prompt: &survey.Select{
				Message: "Type:",
				Options: []string{"pipe", "tcp", "udp", "set"},
				Default: "pipe",
			},
			Validate: survey.Required,
		},
	}

	survey.Ask(qs, opts)
}

func (opts *handlerOpts) queryForCommand() {
	var qs = []*survey.Question{
		{
			Name:     "command",
			Prompt:   &survey.Input{"Command:", ""},
			Validate: survey.Required,
		},
	}

	survey.Ask(qs, opts)
}

func (opts *handlerOpts) queryForSocket() {
	var qs = []*survey.Question{
		{
			Name:     "socketHost",
			Prompt:   &survey.Input{"Socket Host:", ""},
			Validate: survey.Required,
		},
		{
			Name:     "socketPort",
			Prompt:   &survey.Input{"Socket Port:", ""},
			Validate: survey.Required,
		},
	}

	survey.Ask(qs, opts)
}

func (opts *handlerOpts) queryForHandlers() {
	var qs = []*survey.Question{
		{
			Name:     "handlers",
			Prompt:   &survey.Input{"Handlers:", ""},
			Validate: survey.Required,
		},
	}

	survey.Ask(qs, opts)
}

func (opts *handlerOpts) toHandler() *types.Handler {
	handler := &types.Handler{}
	handler.Name = opts.Name
	handler.Type = strings.ToLower(opts.Type)

	if len(opts.Mutator) > 0 {
		handler.Mutator = opts.Mutator
	}

	if len(opts.Command) > 0 {
		handler.Command = opts.Command
	}

	if len(opts.Timeout) > 0 {
		t, _ := strconv.Atoi(opts.Timeout)
		handler.Timeout = t
	}

	if len(opts.SocketHost) > 0 && len(opts.SocketPort) > 0 {
		p, _ := strconv.Atoi(opts.SocketPort)
		handler.Socket = types.HandlerSocket{
			Host: opts.SocketHost,
			Port: p,
		}
	}

	if len(opts.Handlers) > 0 {
		handlers := helpers.SafeSplitCSV(opts.Handlers)
		handler.Handlers = make([]string, len(handlers))
		for i, h := range handlers {
			handler.Handlers[i] = strings.TrimSpace(h)
		}
	}

	return handler
}

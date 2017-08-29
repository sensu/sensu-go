package handler

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// CreateCommand adds command that allows the user to create new handlers
func CreateCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create NAME",
		Short:        "create new handlers",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			flags := cmd.Flags()
			isInteractive := flags.NFlag() == 0

			opts := newHandlerOpts()
			opts.Env = cli.Config.Environment()
			opts.Org = cli.Config.Organization()

			if len(args) > 0 {
				opts.Name = args[0]
			}

			if isInteractive {
				opts.administerQuestionnaire(false)
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

	cmd.Flags().StringP("type", "t", typeDefault, "type of handler (pipe, tcp, udp, or set)")
	cmd.Flags().StringP("mutator", "m", "", "Sensu event mutator (name) to use to mutate event data for the handler")
	cmd.Flags().StringP("command", "c", "", "command to be executed. The event data is passed to the process via STDIN")
	cmd.Flags().StringP("timeout", "i", "", "execution duration timeout in seconds (hard stop)")
	cmd.Flags().String("socket-host", "", "host of handler socket")
	cmd.Flags().String("socket-port", "", "port of handler socket")
	cmd.Flags().StringP("handlers", "", "", "comma separated list of handlers to call")

	return cmd
}

func (opts *handlerOpts) toHandler() *types.Handler {
	handler := &types.Handler{}
	handler.Name = opts.Name
	handler.Environment = opts.Env
	handler.Organization = opts.Org
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

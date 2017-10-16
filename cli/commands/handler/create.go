package handler

import (
	"fmt"

	"github.com/sensu/sensu-go/cli"
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

			if len(args) > 0 {
				opts.Name = args[0]
			}

			if isInteractive {
				if err := opts.administerQuestionnaire(false); err != nil {
					return err
				}
			} else {
				opts.withFlags(flags)
			}

			if opts.Org == "" {
				opts.Org = cli.Config.Organization()
			}

			if opts.Env == "" {
				opts.Env = cli.Config.Environment()
			}

			handler := types.Handler{}
			opts.Copy(&handler)

			if err := handler.Validate(); err != nil {
				if !isInteractive {
					cmd.Help()
				}
				return err
			}

			err := cli.Client.CreateHandler(&handler)
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

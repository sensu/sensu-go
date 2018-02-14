package event

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// ResolveCommand manually resolves an event
func ResolveCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "resolve [ENTITY] [CHECK]",
		Short:        "manually resolves an event",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			entity := args[0]
			check := args[1]

			event, err := cli.Client.FetchEvent(entity, check)
			if err != nil {
				return err
			}

			// Resolve event via api
			if err := cli.Client.ResolveEvent(event); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "OK")
			return nil
		},
	}

	return cmd
}

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
				cmd.Help()
				return errors.New("missing argument(s)")
			}

			// Delete event via API
			entity := args[0]
			check := args[1]

			event, err := cli.Client.FetchEvent(entity, check)
			if err != nil {
				return err
			}

			if err := cli.Client.ResolveEvent(event); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "OK")
			return nil
		},
	}

	cmd.Flags().Bool("skip-confirm", false, "skip interactive confirmation prompt")

	return cmd
}

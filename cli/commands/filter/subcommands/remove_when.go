package subcommands

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// RemoveWhenCommand adds a command that allows a user to remove the time
// windows of a filter
func RemoveWhenCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "remove-when FILTER",
		Short:        "removes time windows from a filter",
		SilenceUsage: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Print usage if we do not receive one argument
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			filter, err := cli.Client.FetchFilter(args[0])
			if err != nil {
				return err
			}
			filter.When = nil

			if err := filter.Validate(); err != nil {
				return err
			}
			if err := cli.Client.UpdateFilter(filter); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Removed")
			return nil
		},
	}

	return cmd
}

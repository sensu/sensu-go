package subcommands

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// RemoveLowFlapThresholdCommand adds a command that allows a user to remove
// the low flap threshold of a check
func RemoveLowFlapThresholdCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "remove-low-flap-threshold [NAME]",
		Short:        "removes low flap threshold from a check",
		SilenceUsage: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Print usage if we do not receive one argument
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			check, err := cli.Client.FetchCheck(args[0])
			if err != nil {
				return err
			}
			check.LowFlapThreshold = 0

			if err := check.Validate(); err != nil {
				return err
			}
			if err := cli.Client.UpdateCheck(check); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "OK")
			return nil
		},
	}

	return cmd
}

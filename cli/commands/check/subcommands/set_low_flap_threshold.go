package subcommands

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// SetLowFlapThresholdCommand updates the low flap threshold of a check
func SetLowFlapThresholdCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "set-low-flap-threshold [NAME] [VALUE]",
		Short:        "set low flap threshold of a check",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			checkName := args[0]
			value := args[1]

			check, err := cli.Client.FetchCheck(checkName)
			if err != nil {
				return err
			}
			lowFlapThreshold, err := strconv.ParseUint(value, 10, 32)
			check.LowFlapThreshold = uint32(lowFlapThreshold)

			if err != nil {
				return err
			}
			if err := check.Validate(); err != nil {
				return err
			}
			if err := cli.Client.UpdateCheck(check); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Updated")
			return nil
		},
	}

	return cmd
}

package subcommands

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// SetRoundRobinCommand updates the round-robin of a check
func SetRoundRobinCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "set-round-robin [NAME] [VALUE]",
		Short:        "set round robin to true or false",
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
			roundRobin, err := strconv.ParseBool(value)
			check.RoundRobin = roundRobin

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

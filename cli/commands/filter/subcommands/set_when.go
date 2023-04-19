package subcommands

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	v2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/timeutil"
	"github.com/spf13/cobra"
)

// SetWhenCommand adds a command that allows a user to set time windows for a
// filter
func SetWhenCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "set-when FILTER",
		Short:        "set time windows for a filter from file or stdin",
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

			timeWindowsPath, _ := cmd.Flags().GetString("file")
			var in *os.File

			if len(timeWindowsPath) > 0 {
				in, err = os.Open(timeWindowsPath)
				if err != nil {
					return err
				}

				defer func() { _ = in.Close() }()
			} else {
				in = os.Stdin
			}
			var timeWindows v2.TimeWindowWhen
			if err := json.NewDecoder(in).Decode(&timeWindows); err != nil {
				return err
			}
			for _, windows := range timeWindows.MapTimeWindows() {
				for _, window := range windows {
					if err := timeutil.ConvertToUTC(window); err != nil {
						return err
					}
				}
			}
			filter.When = &timeWindows
			if err := filter.Validate(); err != nil {
				return err
			}
			if err := cli.Client.UpdateFilter(filter); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Updated")
			return nil
		},
	}

	cmd.Flags().StringP("file", "f", "", "Time windows definition file")

	return cmd
}

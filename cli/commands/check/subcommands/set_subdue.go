package subcommands

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/timeutil"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// SetSubdueCommand adds a command that allows a user to subdue a check
func SetSubdueCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "set-subdue [NAME]",
		Short:        "set subdue of a check from file or stdin",
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

			subduePath, _ := cmd.Flags().GetString("file")
			var in *os.File

			if len(subduePath) > 0 {
				in, err = os.Open(subduePath)
				if err != nil {
					return err
				}

				defer func() { _ = in.Close() }()
			} else {
				in = os.Stdin
			}
			var timeWindows types.TimeWindowWhen
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
			check.Subdue = &timeWindows
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

	cmd.Flags().StringP("file", "f", "", "Subdue definition file")

	return cmd
}

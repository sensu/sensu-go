package check

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// SubdueCommand adds a command that allows a user to subdue a check
func SubdueCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "subdue NAME",
		Short:        "subdue checks from file or stdin",
		SilenceUsage: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Print usage if we do not receive one argument
			if len(args) != 1 {
				return cmd.Help()
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
					if err := helpers.ConvertToUTC(window); err != nil {
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

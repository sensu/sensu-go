package subcommands

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// RemoveCheckHookCommand defines new command to delete hooks from a check
func RemoveCheckHookCommand(cli *cli.SensuCli) *cobra.Command {
	return &cobra.Command{
		Use:          "remove-hook [CHECKNAME] [TYPE] [HOOKNAME]",
		Short:        "removes a hook from a check",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no name is present print out usage
			if len(args) != 3 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			checkName := args[0]
			checkHookType := args[1]
			hookName := args[2]

			check, err := cli.Client.FetchCheck(checkName)
			if err != nil {
				return err
			}

			err = cli.Client.RemoveCheckHook(check, checkHookType, hookName)
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Removed")
			return nil
		},
	}
}

package role

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// RemoveRuleCommand defines new command to delete roles
func RemoveRuleCommand(cli *cli.SensuCli) *cobra.Command {
	return &cobra.Command{
		Use:          "remove-rule [ROLE-NAME] [TYPE]",
		Short:        "remove rule given name",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no name is present print out usage
			if len(args) != 2 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			name := args[0]
			ruleType := args[1]

			err := cli.Client.RemoveRule(name, ruleType)
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Removed")
			return nil
		},
	}
}

package configure

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

func SetOrgCommand(cli *cli.SensuCli) *cobra.Command {
	return &cobra.Command{
		Use:          "set-organization [ORGANIZATION]",
		Short:        "Set organization for active profile",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if argsLen := len(args); argsLen == 0 {
				return errors.New("please provide the name of the organization as an argument")
			} else if argsLen > 1 {
				return errors.New("too many arguments provided")
			}

			newOrg := args[0]
			if err := cli.Config.SaveOrganization(newOrg); err != nil {
				fmt.Fprintf(
					cmd.OutOrStderr(),
					"Unable to write new configuration file with error: %s\n",
					err,
				)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "OK")
			return nil
		},
	}
}

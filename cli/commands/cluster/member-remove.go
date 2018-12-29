package cluster

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// MemberRemoveCommand removes a cluster member by ID
func MemberRemoveCommand(cli *cli.SensuCli) *cobra.Command {
	return &cobra.Command{
		Use:          "member-remove [ID]",
		Short:        "remove cluster member by ID",
		SilenceUsage: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			memberID := args[0]
			id, err := strconv.ParseUint(memberID, 16, 64)
			if err != nil {
				return fmt.Errorf("invalid id: %s", err)
			}

			if _, err := cli.Client.MemberRemove(id); err != nil {
				return fmt.Errorf("error removing cluster member: %s", err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "OK")
			return nil
		},
	}
}

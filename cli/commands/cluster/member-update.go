package cluster

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// MemberUpdateCommand updates a cluster member
func MemberUpdateCommand(cli *cli.SensuCli) *cobra.Command {
	return &cobra.Command{
		Use:          "member-update [ID] [PEER_ADDRS]",
		Short:        "update cluster member by ID with comma-separated peer addresses",
		SilenceUsage: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			memberID := args[0]
			peerAddrs := splitAndTrim(args[1])

			id, err := strconv.ParseUint(memberID, 16, 64)
			if err != nil {
				return fmt.Errorf("invalid id: %s", err)
			}

			if _, err := cli.Client.MemberUpdate(id, peerAddrs); err != nil {
				return fmt.Errorf("error updating cluster member: %s", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Updated member with ID %x in cluster\n", id)
			return nil
		},
	}
}

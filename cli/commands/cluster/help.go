package cluster

import (
	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// HelpCommand defines new parent
func HelpCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "manage sensu cluster",
	}

	cmd.AddCommand(
		MemberListCommand(cli),
		MemberAddCommand(cli),
	)

	return cmd
}

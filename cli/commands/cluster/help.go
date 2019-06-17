package cluster

import (
	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// HelpCommand defines new parent
func HelpCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Manage sensu cluster",
	}

	cmd.AddCommand(
		MemberListCommand(cli),
		MemberAddCommand(cli),
		MemberUpdateCommand(cli),
		MemberRemoveCommand(cli),
		HealthCommand(cli),
		IDCommand(cli),
	)

	return cmd
}

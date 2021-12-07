package opampagent

import (
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/spf13/cobra"
)

func HelpCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "opampagent",
		Short: "Manage global open-telemetetry collector configuration",
		RunE:  helpers.DefaultSubCommandRunE,
	}
	cmd.AddCommand(
		ConfigureCommand(cli),
		ListConfigCommand(cli),
	)
	return cmd
}

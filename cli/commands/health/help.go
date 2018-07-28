package health

import (
	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// HelpCommand defines new parent
func HelpCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "health",
		Short: "view sensu health information",
	}
	/*
		cmd.AddCommand(
			HealthCommand(cli),
		)
	*/
	return cmd
}

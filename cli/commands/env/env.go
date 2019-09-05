package env

import (
	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// Command display the commands to set up the environment used by sensuctl
func Command(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "env",
		Short: "display the commands to set up the environment used by sensuctl",
		RunE:  execute(cli),
	}

	return cmd
}

func execute(cli *cli.SensuCli) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return nil
	}
}

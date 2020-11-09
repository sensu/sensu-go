package silenced

import (
	"time"

	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

const (
	timeFormat = time.RFC3339
)

// HelpCommand displays help
func HelpCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "silenced",
		Short: "Manage silenced subscriptions and checks",
	}

	// Add sub-commands
	cmd.AddCommand(
		CreateCommand(cli),
		DeleteCommand(cli),
		ListCommand(cli),
		InfoCommand(cli),
		UpdateCommand(cli),
	)

	return cmd
}

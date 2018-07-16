package licensing

import (
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// HelpCommand defines new parent
func HelpCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "license",
		Short:       "Manage enterprise license",
		Hidden:      !cli.Config.IsEnterprise(),
		Annotations: map[string]string{"edition": types.EnterpriseEdition},
	}

	// Add sub-commands
	cmd.AddCommand(
		InfoCommand(cli),
	)
	return cmd
}

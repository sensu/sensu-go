package check

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// ListCommand defines new list events command
func ListCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "list",
		Short:        "list checks",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := cli.Client.ListChecks()
			if err != nil {
				return err
			}

			result, _ := json.MarshalIndent(r, "", "  ")
			fmt.Fprintf(os.Stdout, "%s\n", result)
			return nil
		},
	}

	return cmd
}

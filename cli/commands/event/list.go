package event

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
		Use:   "list",
		Short: "list events",
		Run: func(cmd *cobra.Command, args []string) {
			r, err := cli.Client.ListEvents()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err)
				os.Exit(1)
			}

			result, _ := json.Marshal(r)
			fmt.Fprintf(os.Stdout, "%s\n", result)
		},
	}

	return cmd
}

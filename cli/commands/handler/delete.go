package handler

import (
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// DeleteCommand adds a command that allows user to delete handlers
func DeleteCommand(cli *cli.SensuCli) *cobra.Command {
	return &cobra.Command{
		Use:          "delete [NAME]",
		Short:        "delete handlers given name",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no name is present print out usage
			if len(args) != 1 {
				cmd.Help()
				return nil
			}

			name := args[0]
			handler := &types.Handler{Name: name}
			err := cli.Client.DeleteHandler(handler)
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "OK")
			return nil
		},
	}
}

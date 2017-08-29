package handler

import (
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// UpdateCommand allows the user to update handlers
func UpdateCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "update NAME",
		Short:        "update handlers",
		SilenceUsage: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			//Fetch handlers from API
			handlerName := args[0]
			handler, err := cli.Client.FetchHandler(handlerName)
			if err != nil {
				return err
			}

			opts := newHandlerOpts()
			opts.withHandler(handler)

			opts.administerQuestionnaire(true)

			opts.Copy(handler)

			if err := handler.Validate(); err != nil {
				return err
			}

			if err := cli.Client.CreateHandler(handler); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "OK")
			return nil
		},
	}

	return cmd
}

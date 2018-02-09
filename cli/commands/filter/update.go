package filter

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// UpdateCommand defines the 'filter update' subcommand
func UpdateCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "update [NAME]",
		Short:        "update filters",
		SilenceUsage: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Print out usage if we do not receive one argument
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			// Fetch the requested filter from the API
			name := args[0]
			filter, err := cli.Client.FetchFilter(name)
			if err != nil {
				return err
			}

			// Administer questionnaire
			opts := newFilterOpts()
			opts.withFilter(filter)
			if err := opts.administerQuestionnaire(true); err != nil {
				return err
			}

			// Apply given arguments to check
			opts.Copy(filter)

			if err := filter.Validate(); err != nil {
				return err
			}

			if err := cli.Client.UpdateFilter(filter); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "OK")
			return nil
		},
	}

	return cmd
}

package check

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// UpdateCommand adds command that allows user to create new checks
func UpdateCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "update [NAME]",
		Short:        "update checks",
		SilenceUsage: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Print usage if we do not receive one argument
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			// Fetch checks from API
			checkID := args[0]
			check, err := cli.Client.FetchCheck(checkID)
			if err != nil {
				return err
			}

			// Administer questionnaire
			opts := newCheckOpts()
			opts.withCheck(check)
			if err := opts.administerQuestionnaire(true); err != nil {
				return err
			}

			// Apply given arguments to check
			opts.Copy(check)

			if err := check.Validate(); err != nil {
				return err
			}

			//
			// TODO:
			//
			// Current validation is a bit too laissez faire. For usability we should
			// determine whether there are assets / handlers / mutators associated w/
			// the check and warn the user if they do not exist yet.
			if err := cli.Client.UpdateCheck(check); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "OK")
			return nil
		},
	}

	return cmd
}

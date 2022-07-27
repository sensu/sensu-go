package mutator

import (
	"errors"
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// UpdateCommand defines the 'mutator update' subcommand
func UpdateCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "update [NAME]",
		Short:        "update mutators",
		SilenceUsage: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Print out usage if we do not receive one argument
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			// Fetch the requested mutator from the API
			name := args[0]
			mutator, err := cli.Client.FetchMutator(name)
			if err != nil {
				return err
			}

			// Administer questionnaire
			opts := newMutatorOpts()
			opts.withMutator(mutator)
			askOpts := []survey.AskOpt{
				survey.WithStdio(cli.InFile, cli.OutFile, cli.ErrFile),
			}
			if err := opts.administerQuestionnaire(true, askOpts...); err != nil {
				return err
			}

			// Apply given arguments to mutator
			opts.Copy(mutator)

			if err := mutator.Validate(); err != nil {
				return err
			}

			if err := cli.Client.UpdateMutator(mutator); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Updated")
			return nil
		},
	}

	return cmd
}

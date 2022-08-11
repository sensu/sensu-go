package entity

import (
	"errors"
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// UpdateCommand adds command that allows user to create new checks
func UpdateCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "update [ID]",
		Short:        "update entity",
		SilenceUsage: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Print out usage if we do not receive one argument
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			// Fetch the specified entity from API
			id := args[0]
			entity, err := cli.Client.FetchEntity(id)
			if err != nil {
				return err
			}

			// Administer questionnaire
			opts := newEntityOpts()
			opts.withEntity(entity)
			askOpts := []survey.AskOpt{
				survey.WithStdio(cli.InFile, cli.OutFile, cli.ErrFile),
			}
			if err := opts.administerQuestionnaire(true, askOpts...); err != nil {
				return err
			}

			// Apply given arguments to check
			opts.copy(entity)

			if err := entity.Validate(); err != nil {
				return err
			}

			if err := cli.Client.UpdateEntity(entity); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Updated")
			return nil
		},
	}

	return cmd
}

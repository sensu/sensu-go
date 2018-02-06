package entity

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// DeleteCommand adds a command that allows user to delete entities
func DeleteCommand(cli *cli.SensuCli) *cobra.Command {
	exec := &deleteExecutor{client: cli.Client}
	cmd := &cobra.Command{
		Use:          "delete [ID]",
		Short:        "delete entity given ID",
		RunE:         exec.run,
		SilenceUsage: true,
	}

	cmd.Flags().Bool("skip-confirm", false, "skip interactive confirmation prompt")

	return cmd
}

type deleteExecutor struct {
	client client.APIClient
}

func (e *deleteExecutor) run(cmd *cobra.Command, args []string) error {
	// If no ID was given print out usage
	id, err := e.extractID(args)
	if err != nil {
		return cmd.Help()
	}

	if skipConfirm, _ := cmd.Flags().GetBool("skip-confirm"); !skipConfirm {
		if confirmed := helpers.ConfirmDelete(id); !confirmed {
			fmt.Fprintln(cmd.OutOrStdout(), "Canceled")
			return nil
		}
	}

	if err := e.deleteEntityByID(id); err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), "OK")
	return nil
}

func (e *deleteExecutor) extractID(args []string) (string, error) {
	// If no name is present print out usage
	if len(args) < 1 {
		return "", errors.New("name argument not received")
	} else if len(args) > 1 {
		return args[0], errors.New("too many arguments received")
	}

	return args[0], nil
}

func (e *deleteExecutor) deleteEntityByID(id string) (err error) {
	entity := &types.Entity{ID: id}
	err = e.client.DeleteEntity(entity)

	return
}

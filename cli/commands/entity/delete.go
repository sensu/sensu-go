package entity

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/spf13/cobra"
)

// DeleteCommand adds a command that allows user to delete entities
func DeleteCommand(cli *cli.SensuCli) *cobra.Command {
	exec := &deleteExecutor{client: cli.Client}
	cmd := &cobra.Command{
		Use:          "delete [NAME]",
		Short:        "delete entity given name",
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
	// If no Name was given print out usage
	if len(args) != 1 {
		_ = cmd.Help()
		return errors.New("invalid argument(s) received")
	}
	name := args[0]

	if skipConfirm, _ := cmd.Flags().GetBool("skip-confirm"); !skipConfirm {
		if confirmed := helpers.ConfirmDelete(name); !confirmed {
			fmt.Fprintln(cmd.OutOrStdout(), "Canceled")
			return nil
		}
	}

	if err := e.deleteEntityByName(name); err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), "OK")
	return nil
}

func (e *deleteExecutor) deleteEntityByName(name string) (err error) {
	entity := &v2.Entity{ObjectMeta: v2.NewObjectMeta(name, "")}
	return e.client.DeleteEntity(entity)
}

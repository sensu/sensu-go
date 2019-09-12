package command

import (
	"errors"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/cmdmanager"
	"github.com/spf13/cobra"
)

// ExecCommand provides a method to execute command plugins.
func ExecCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exec [ALIAS] [args]",
		Short: "executes a sensuctl command plugin",
		RunE:  execCommandExecute(cli),
	}

	return cmd
}

func execCommandExecute(cli *cli.SensuCli) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// If no name is present print out usage
		if len(args) < 1 {
			_ = cmd.Help()
			return errors.New("invalid argument(s) received")
		}

		manager, err := cmdmanager.NewCommandManager()
		if err != nil {
			return err
		}

		if err = manager.ExecCommand(args[0], args[1:]); err != nil {
			return err
		}

		return nil
	}
}

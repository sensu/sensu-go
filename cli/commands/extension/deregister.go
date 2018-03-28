package extension

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client"
	"github.com/spf13/cobra"
)

// DeregisterCommand adds command that allows deregistering extensions
func DeregisterCommand(cli *cli.SensuCli) *cobra.Command {
	return &cobra.Command{
		Use:   "deregister NAME",
		Short: "deregister extensions",
		RunE:  runDeregister(cli.Client, cli.Config.Organization()),
	}
}

func runDeregister(client client.APIClient, org string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("invalid arguments received")
		}
		name := args[0]
		if err := client.DeregisterExtension(name, org); err != nil {
			return err
		}

		fmt.Fprintln(cmd.OutOrStdout(), "OK")
		return nil
	}
}

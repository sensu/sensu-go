package extension

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// RegisterCommand adds command that allows registering extensions
func RegisterCommand(cli *cli.SensuCli) *cobra.Command {
	return &cobra.Command{
		Use:   "register NAME URL",
		Short: "register extensions",
		RunE:  runRegister(cli.Client, cli.Config.Organization()),
	}
}

func runRegister(client client.APIClient, org string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			_ = cmd.Help()
			return errors.New("invalid arguments received")
		}

		name := args[0]
		url := args[1]

		extension := types.Extension{
			Organization: org,
			Name:         name,
			URL:          url,
		}

		if err := extension.Validate(); err != nil {
			return err
		}

		if err := client.RegisterExtension(&extension); err != nil {
			return err
		}

		fmt.Fprintln(cmd.OutOrStdout(), "OK")
		return nil
	}
}

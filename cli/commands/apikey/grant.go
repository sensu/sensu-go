package apikey

import (
	"errors"
	"fmt"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// GrantCommand adds a command that creates apikeys.
func GrantCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "grant [USERNAME]",
		Short:        "grant new apikey",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			apikey := &corev2.APIKey{
				Username: args[0],
			}

			if err := cli.Client.Post(apikey.URIPath(), apikey); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Created")
			return nil
		},
	}

	return cmd
}

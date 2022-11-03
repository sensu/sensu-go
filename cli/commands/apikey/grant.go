package apikey

import (
	"errors"
	"fmt"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// GrantCommand adds a command that creates apikeys.
func GrantCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "grant [USERNAME]",
		Short:        "grant new api-key",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			apikey := &corev2.APIKey{
				Username: args[0],
			}

			location, err := cli.Client.PostAPIKey(apikey.URIPath(), apikey)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Created: %s\n", location)
			return nil
		},
	}

	return cmd
}

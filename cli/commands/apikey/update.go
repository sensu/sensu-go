package apikey

import (
	"errors"
	"fmt"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// UpdateCommand adds a command that updates apikeys.
func UpdateCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "update [NAME] [USERNAME]",
		Short:        "update apikeys",
		SilenceUsage: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			apikey := &corev2.APIKey{
				ObjectMeta: corev2.ObjectMeta{
					Name: args[0],
				},
				Username: args[1],
			}

			err := cli.Client.Patch(apikey.URIPath(), apikey)
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Updated")
			return nil
		},
	}

	return cmd
}

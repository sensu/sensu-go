package silenced

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// UpdateCommand updates a given silenced entry
func UpdateCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "update [ID]",
		Short:        "update silenced entries",
		SilenceUsage: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}
			id, err := getID(cmd, args)
			if err != nil {
				return err
			}

			silenced, err := cli.Client.FetchSilenced(id)
			if err != nil {
				return err
			}

			opts := toOpts(silenced)

			if err := opts.administerQuestionnaire(true); err != nil {
				return err
			}

			if err := opts.Apply(silenced); err != nil {
				return err
			}

			if err := silenced.Validate(); err != nil {
				return err
			}

			if err := cli.Client.UpdateSilenced(silenced); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "OK")
			return nil
		},
	}
	_ = cmd.Flags().StringP("subscription", "s", "", "silenced subscription")
	_ = cmd.Flags().StringP("check", "c", "", "silenced check")
	_ = cmd.Flags().BoolP("expire-on-resolve", "x", false, "clear silenced entry on resolution")
	_ = cmd.Flags().StringP("expire", "e", expireDefault, "expiry in seconds")
	_ = cmd.Flags().StringP("begin", "b", beginDefault, "silence begin in human readable time (Format: Jan 02 2006 3:04PM MST)")
	return cmd
}

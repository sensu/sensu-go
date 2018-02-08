package silenced

import (
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// CreateCommand is a command that creates new silenceds
func CreateCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create",
		Short:        "create a silenced entry",
		SilenceUsage: true,
		Args:         cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			flags := cmd.Flags()
			isInteractive := flags.NFlag() == 0

			opts := newSilencedOpts()

			opts.Org = cli.Config.Organization()
			opts.Env = cli.Config.Environment()

			if isInteractive {
				if err := opts.administerQuestionnaire(false); err != nil {
					return err
				}
			} else {
				if err := opts.withFlags(flags); err != nil {
					return err
				}
				if opts.Check == "" && opts.Subscription == "" {
					return fmt.Errorf("must specify --check or --subscription")
				}
				if opts.Reason == "" {
					return fmt.Errorf("must specify --reason")
				}
			}
			var silenced types.Silenced
			if err := opts.Apply(&silenced); err != nil {
				return err
			}
			if err := silenced.Validate(); err != nil {
				return err
			}
			if err := cli.Client.CreateSilenced(&silenced); err != nil {
				return err
			}

			_, err := fmt.Fprintln(cmd.OutOrStdout(), "OK")
			return err
		},
	}

	_ = cmd.Flags().StringP("reason", "r", "", "reason for the silenced entry")
	_ = cmd.Flags().BoolP("expire-on-resolve", "x", false, "clear silenced entry on resolution")
	_ = cmd.Flags().StringP("expire", "e", expireDefault, "expiry in seconds")
	_ = cmd.Flags().StringP("subscription", "s", "", "silence subscription")
	_ = cmd.Flags().StringP("check", "c", "", "silence check")
	_ = cmd.Flags().StringP("begin", "b", beginDefault, "silence begin in human readable time (Format: Jan 02 2006 3:04PM MST)")

	return cmd
}

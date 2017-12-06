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
		RunE: func(cmd *cobra.Command, args []string) error {
			flags := cmd.Flags()
			isInteractive := flags.NFlag() == 0

			opts := newSilencedOpts()

			if len(args) > 0 {
				opts.Subscription = args[0]
			}
			if len(args) > 1 {
				opts.Check = args[1]
			}

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

	cmd.Flags().StringP("reason", "r", "", "reason for the silenced entry")
	cmd.Flags().BoolP("expire-on-resolve", "x", false, "clear silenced entry on resolution")
	cmd.Flags().Int64P("expire", "e", 0, "expiry in seconds")
	cmd.Flags().StringP("subscription", "s", "", "silence subscription")
	cmd.Flags().StringP("check", "c", "", "silence check")

	cmd.MarkFlagRequired("reason")

	return cmd
}

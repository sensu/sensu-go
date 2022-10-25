package tessen

import (
	"errors"
	"fmt"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/spf13/cobra"
)

// OptOutCommand updates the tessen configuration to opt-out
func OptOutCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "opt-out",
		Short:        "opt-out to the tessen call home service",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			if skipConfirm, _ := cmd.Flags().GetBool("skip-confirm"); !skipConfirm {
				if confirmed := helpers.ConfirmOptOut(); !confirmed {
					fmt.Fprintln(cmd.OutOrStdout(), "Phew! We're glad you decided to stick around for a while!")
					return nil
				}
			}

			config := corev2.TessenConfig{OptOut: true}
			if err := cli.Client.Put(config.URIPath(), config); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "I thought we were friends :( Remember, you can opt back in at any time!")
			return nil
		},
	}

	cmd.Flags().Bool("skip-confirm", false, "skip interactive confirmation prompt")

	return cmd
}

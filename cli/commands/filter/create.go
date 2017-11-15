package filter

import (
	"fmt"
	"strings"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// CreateCommand defines the 'filter create' subcommand
func CreateCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create NAME",
		Short:        "create new filters",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			flags := cmd.Flags()
			isInteractive := flags.NFlag() == 0

			opts := newFilterOpts()

			if len(args) > 0 {
				opts.Name = args[0]
			}

			opts.Org = cli.Config.Organization()
			opts.Env = cli.Config.Environment()

			if isInteractive {
				if err := opts.administerQuestionnaire(false); err != nil {
					return err
				}
			} else {
				opts.withFlags(flags)
			}

			// Apply given arguments to check
			filter := types.EventFilter{}
			opts.Copy(&filter)

			if err := filter.Validate(); err != nil {
				if !isInteractive {
					cmd.SilenceUsage = false
				}
				return err
			}

			if err := cli.Client.CreateFilter(&filter); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "OK")
			return nil
		},
	}

	cmd.Flags().StringP("action", "a", "",
		"specifies whether events are passed through the filter or blocked by the "+
			"filter. Allowed values: "+strings.Join(types.EventFilterAllActions, ", "),
	)
	cmd.Flags().StringP("statements", "s", "",
		"comma separated list of boolean expressions that are evaluated to "+
			"determine if the event matches this filter",
	)

	// Mark flags are required for bash-completions
	cmd.MarkFlagRequired("action")
	cmd.MarkFlagRequired("statements")

	return cmd
}

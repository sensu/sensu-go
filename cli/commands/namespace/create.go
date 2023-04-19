package namespace

import (
	"errors"
	"fmt"

	v2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/flags"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/spf13/cobra"
)

// CreateCommand adds command that allows users to create new namespaces
func CreateCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create [NAME]",
		Short:        "create new namespace",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			isInteractive, _ := cmd.Flags().GetBool(flags.Interactive)
			opts := newNamespaceOpts()

			if len(args) > 0 {
				opts.Name = args[0]
			}

			if isInteractive {
				if err := opts.administerQuestionnaire(false); err != nil {
					return err
				}
			}

			namespace := v2.Namespace{}
			opts.Copy(&namespace)

			if err := namespace.Validate(); err != nil {
				if !isInteractive {
					cmd.SilenceUsage = false
				}
				return err
			}

			if err := cli.Client.CreateNamespace(&namespace); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Created")
			return nil
		},
	}

	helpers.AddInteractiveFlag(cmd.Flags())
	return cmd
}

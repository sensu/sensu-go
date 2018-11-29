package entity

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/flags"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/spf13/cobra"
)

// CreateCommand allows a user to create a new entity
func CreateCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create [NAME]",
		Short:        "create a new entity",
		SilenceUsage: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			isInteractive, _ := cmd.Flags().GetBool(flags.Interactive)
			if !isInteractive {
				// Mark flags are required for bash-completions
				_ = cmd.MarkFlagRequired("name")
				_ = cmd.MarkFlagRequired("entity-class")
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			isInteractive, _ := cmd.Flags().GetBool(flags.Interactive)
			opts := newEntityOpts()

			if len(args) > 0 {
				opts.Name = args[0]
			}

			opts.Namespace = cli.Config.Namespace()

			if isInteractive {
				if err := opts.administerQuestionnaire(); err != nil {
					return err
				}
			} else {
				opts.withFlags(cmd.Flags())
			}

			// Apply given arguments to entity
			entity := v2.NewEntity(v2.NewObjectMeta("", ""))
			opts.copy(entity)

			if err := entity.Validate(); err != nil {
				if !isInteractive {
					_ = cmd.Help()
				}
				return err
			}

			if err := cli.Client.CreateEntity(entity); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "OK")
			return nil
		},
	}

	_ = cmd.Flags().StringP("entity-class", "c", "", "entity class, either proxy or agent")
	_ = cmd.Flags().StringP("subscriptions", "s", "", "comma separated list of subscriptions")
	helpers.AddInteractiveFlag(cmd.Flags())
	return cmd

}

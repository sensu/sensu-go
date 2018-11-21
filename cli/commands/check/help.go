package check

import (
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/check/subcommands"
	"github.com/spf13/cobra"
)

// HelpCommand defines new parent
func HelpCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Manage checks",
	}

	// Add sub-commands
	cmd.AddCommand(
		// CRUD Commands
		CreateCommand(cli),
		DeleteCommand(cli),
		ExecuteCommand(cli),
		ListCommand(cli),
		InfoCommand(cli),
		UpdateCommand(cli),

		// Remove commands (clear out fields)
		subcommands.RemoveCheckHookCommand(cli),
		// cannot remove command, required field
		// cannot remove cron, use set-interval
		subcommands.RemoveHandlersCommand(cli),
		subcommands.RemoveHighFlapThresholdCommand(cli),
		// cannot remove interval, use set-cron
		subcommands.RemoveLowFlapThresholdCommand(cli),
		subcommands.RemoveProxyEntityNameCommand(cli),
		subcommands.RemoveProxyRequestsCommand(cli),
		// cannot remove publish, use set-publish
		subcommands.RemoveRuntimeAssetsCommand(cli),
		// cannot remove stdin, use set-stdin

		// TODO: re-enable this after fixing check subdue
		// subcommands.RemoveSubdueCommand(cli),

		// cannot remove subscriptions, required field
		subcommands.RemoveTTLCommand(cli),
		subcommands.RemoveTimeoutCommand(cli),
		subcommands.RemoveOutputMetricHandlersCommand(cli),
		subcommands.RemoveOutputMetricFormatCommand(cli),

		// Set commands (update fields)
		subcommands.SetCheckHooksCommand(cli),
		subcommands.SetCommandCommand(cli),
		subcommands.SetCronCommand(cli),
		subcommands.SetHandlersCommand(cli),
		subcommands.SetHighFlapThresholdCommand(cli),
		subcommands.SetIntervalCommand(cli),
		subcommands.SetLowFlapThresholdCommand(cli),
		subcommands.SetProxyEntityNameCommand(cli),
		subcommands.SetProxyRequestsCommand(cli),
		subcommands.SetPublishCommand(cli),
		subcommands.SetRuntimeAssetsCommand(cli),
		subcommands.SetSTDINCommand(cli),

		// TODO: re-enable this after fixing check subdue
		// subcommands.SetSubdueCommand(cli),

		subcommands.SetSubscriptionsCommand(cli),
		subcommands.SetTTLCommand(cli),
		subcommands.SetTimeoutCommand(cli),
		subcommands.SetOutputMetricHandlersCommand(cli),
		subcommands.SetOutputMetricFormatCommand(cli),
		subcommands.SetRoundRobinCommand(cli),
	)

	return cmd
}

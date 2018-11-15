package commands

import (
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/asset"
	"github.com/sensu/sensu-go/cli/commands/check"
	"github.com/sensu/sensu-go/cli/commands/cluster"
	"github.com/sensu/sensu-go/cli/commands/completion"
	"github.com/sensu/sensu-go/cli/commands/config"
	"github.com/sensu/sensu-go/cli/commands/configure"
	"github.com/sensu/sensu-go/cli/commands/create"
	"github.com/sensu/sensu-go/cli/commands/edit"
	"github.com/sensu/sensu-go/cli/commands/entity"
	"github.com/sensu/sensu-go/cli/commands/event"
	"github.com/sensu/sensu-go/cli/commands/extension"
	"github.com/sensu/sensu-go/cli/commands/filter"
	"github.com/sensu/sensu-go/cli/commands/handler"
	"github.com/sensu/sensu-go/cli/commands/hook"
	"github.com/sensu/sensu-go/cli/commands/logout"
	"github.com/sensu/sensu-go/cli/commands/mutator"
	"github.com/sensu/sensu-go/cli/commands/namespace"
	"github.com/sensu/sensu-go/cli/commands/role"
	"github.com/sensu/sensu-go/cli/commands/silenced"
	"github.com/sensu/sensu-go/cli/commands/user"
	"github.com/spf13/cobra"
)

// AddCommands adds management commands to given command
func AddCommands(rootCmd *cobra.Command, cli *cli.SensuCli) {
	rootCmd.AddCommand(
		configure.Command(cli),
		completion.Command(rootCmd),
		logout.Command(cli),

		// Management Commands
		asset.HelpCommand(cli),
		check.HelpCommand(cli),
		config.HelpCommand(cli),
		entity.HelpCommand(cli),
		event.HelpCommand(cli),
		filter.HelpCommand(cli),
		handler.HelpCommand(cli),
		hook.HelpCommand(cli),
		mutator.HelpCommand(cli),
		namespace.HelpCommand(cli),
		role.HelpCommand(cli),
		user.HelpCommand(cli),
		silenced.HelpCommand(cli),
		create.CreateCommand(cli),
		extension.HelpCommand(cli),
		cluster.HelpCommand(cli),
		edit.Command(cli),
	)

	for _, cmd := range rootCmd.Commands() {
		rootCmd.ValidArgs = append(rootCmd.ValidArgs, cmd.Use)
	}
}

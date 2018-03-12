package commands

import (
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/asset"
	"github.com/sensu/sensu-go/cli/commands/check"
	"github.com/sensu/sensu-go/cli/commands/completion"
	"github.com/sensu/sensu-go/cli/commands/config"
	"github.com/sensu/sensu-go/cli/commands/configure"
	"github.com/sensu/sensu-go/cli/commands/entity"
	"github.com/sensu/sensu-go/cli/commands/environment"
	"github.com/sensu/sensu-go/cli/commands/event"
	"github.com/sensu/sensu-go/cli/commands/filter"
	"github.com/sensu/sensu-go/cli/commands/handler"
	"github.com/sensu/sensu-go/cli/commands/hook"
	"github.com/sensu/sensu-go/cli/commands/importer"
	"github.com/sensu/sensu-go/cli/commands/logout"
	"github.com/sensu/sensu-go/cli/commands/mutator"
	"github.com/sensu/sensu-go/cli/commands/organization"
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
		importer.ImportCommand(cli),

		// Management Commands
		asset.HelpCommand(cli),
		check.HelpCommand(cli),
		config.HelpCommand(cli),
		entity.HelpCommand(cli),
		environment.HelpCommand(cli),
		event.HelpCommand(cli),
		filter.HelpCommand(cli),
		handler.HelpCommand(cli),
		hook.HelpCommand(cli),
		mutator.HelpCommand(cli),
		organization.HelpCommand(cli),
		role.HelpCommand(cli),
		user.HelpCommand(cli),
		silenced.HelpCommand(cli),
	)

	for _, cmd := range rootCmd.Commands() {
		rootCmd.ValidArgs = append(rootCmd.ValidArgs, cmd.Use)
	}
}

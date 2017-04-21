package commands

import (
	"github.com/sensu/sensu-go/cli/cmd/commands/event"
	"github.com/spf13/cobra"
)

func AddCommands(rootCmd *cobra.Command) {
	cmd.AddCommand(
		// events
		event.NewEventCommand(rootCmd),
	)
}

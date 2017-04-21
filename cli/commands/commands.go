package commands

import (
	"github.com/sensu/sensu-go/cli/commands/event"
	"github.com/spf13/cobra"
)

func AddCommands(rootCmd *cobra.Command) {
	rootCmd.AddCommand(
		// events
		event.NewEventCommand(rootCmd),
	)
}

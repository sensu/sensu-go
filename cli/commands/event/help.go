package event

import "github.com/spf13/cobra"

func NewEventCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "event",
		Short: "Manage events",
	}

	cmd.AddCommand(NewEventListCommand())

	return cmd
}

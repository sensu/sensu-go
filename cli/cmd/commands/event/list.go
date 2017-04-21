package event

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewEventListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list events",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("LISTING THE EVENTS\n", args)
		},
	}

	return cmd
}

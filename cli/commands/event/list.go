package event

import (
	"fmt"

	"github.com/sensu/sensu-go/cli/client"
	"github.com/spf13/cobra"
)

func NewEventListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list events",
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := client.Request().Get("/events")
			if err != nil {
				return nil
			}

			fmt.Printf("LISTING THE EVENTS\n %s", r.String())
			return nil
		},
	}

	return cmd
}

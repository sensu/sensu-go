package event

import (
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

func ListCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list events",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			r, err := cli.Client.R().Get("/events")
			if err != nil {
				return err
			}

			fmt.Println(r.String())
			return
		},
	}

	return cmd
}

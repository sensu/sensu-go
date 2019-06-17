package cluster

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// IDCommand provides the sensu cluster id
func IDCommand(cli *cli.SensuCli) *cobra.Command {
	return &cobra.Command{
		Use:          "id",
		Short:        "show sensu cluster id",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}
			id, err := cli.Client.FetchClusterID()
			if err != nil {
				return err
			}
			fmt.Printf("sensu cluster id: %s", id)
			return nil
		},
	}
}

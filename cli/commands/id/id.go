package id

import (
	"errors"
	"fmt"
	"strings"

	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// Command provides the sensu cluster id
func Command(cli *cli.SensuCli) *cobra.Command {
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
			fmt.Printf("sensu cluster id: %s\n", strings.Trim(id, "\""))
			return nil
		},
	}
}

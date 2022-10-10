package tessen

import (
	"errors"
	"fmt"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// OptInCommand updates the tessen configuration to opt-in
func OptInCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "opt-in",
		Short:        "opt-in to the tessen call home service",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			config := corev2.TessenConfig{OptOut: false}
			if err := cli.Client.Put(config.URIPath(), config); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Thank you so much for opting in!")
			return nil
		},
	}

	return cmd
}

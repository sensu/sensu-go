package silenced

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

// UpdateCommand updates a given silenced entry
func UpdateCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "update [NAME]",
		Short:        "update silenced entries",
		SilenceUsage: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}
			name, err := getName(cmd, args)
			if err != nil {
				return err
			}

			silenced, err := cli.Client.FetchSilenced(name)
			if err != nil {
				return err
			}

			opts := toOpts(silenced)

			if err := opts.administerQuestionnaire(true); err != nil {
				return err
			}

			if err := opts.Apply(silenced); err != nil {
				return err
			}

			if err := silenced.Validate(); err != nil {
				return err
			}

			if err := cli.Client.UpdateSilenced(silenced); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Updated")
			return nil
		},
	}

	return cmd
}

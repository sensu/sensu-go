package role

import (
	"errors"
	"io"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/spf13/cobra"
)

// InfoCommand defines new command to list rules associated w/ a role
func InfoCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "info [ROLE]",
		Aliases:      []string{"list-rules"}, // backward compatibility
		Short:        "show detailed role information",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			// Fetch roles from API
			r, err := cli.Client.FetchRole(args[0])
			if err != nil {
				return err
			}

			// Determine the format to use to output the data
			flag := helpers.GetChangedStringValueFlag("format", cmd.Flags())
			format := cli.Config.Format()
			return helpers.PrintFormatted(flag, format, r, cmd.OutOrStdout(), printRulesToTable)
		},
	}

	helpers.AddFormatFlag(cmd.Flags())

	return cmd
}

func printRulesToTable(v interface{}, io io.Writer) error {
	return nil
}

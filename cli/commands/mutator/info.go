package mutator

import (
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/list"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// InfoCommand defines the 'mutator info' subcommand
func InfoCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "info [NAME]",
		Short:        "show detailed mutator information",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			// Fetch the mutator from API
			name := args[0]
			r, err := cli.Client.FetchMutator(name)
			if err != nil {
				return err
			}

			// Determine the format to use to output the data
			flag := helpers.GetChangedStringValueFlag("format", cmd.Flags())
			format := cli.Config.Format()
			return helpers.PrintFormatted(flag, format, r, cmd.OutOrStdout(), printToList)
		},
	}

	helpers.AddFormatFlag(cmd.Flags())

	return cmd
}

func printToList(v interface{}, writer io.Writer) error {
	mutator, ok := v.(*types.Mutator)
	if !ok {
		return fmt.Errorf("%t is not a Mutator", v)
	}
	cfg := &list.Config{
		Title: mutator.Name,
		Rows: []*list.Row{
			{
				Label: "Name",
				Value: mutator.Name,
			},
			{
				Label: "Command",
				Value: mutator.Command,
			},
			{
				Label: "Timeout",
				Value: strconv.FormatUint(uint64(mutator.Timeout), 10),
			},
			{
				Label: "Namespace",
				Value: mutator.Namespace,
			},
		},
	}

	return list.Print(writer, cfg)
}

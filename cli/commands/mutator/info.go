package mutator

import (
	"errors"
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
			format, _ := cmd.Flags().GetString("format")

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

			if format == "json" {
				return helpers.PrintJSON(r, cmd.OutOrStdout())
			}
			printToList(r, cmd.OutOrStdout())
			return nil
		},
	}

	helpers.AddFormatFlag(cmd.Flags())

	return cmd
}

func printToList(mutator *types.Mutator, writer io.Writer) {
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
				Label: "Organization",
				Value: mutator.Organization,
			},
			{
				Label: "Environment",
				Value: mutator.Environment,
			},
		},
	}

	list.Print(writer, cfg)
}

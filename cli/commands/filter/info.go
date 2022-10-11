package filter

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/list"
	corev2 "github.com/sensu/core/v2"
	"github.com/spf13/cobra"
)

// InfoCommand defines the 'filter info' subcommand
func InfoCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "info [NAME]",
		Short:        "show detailed filter information",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			// Fetch the filter from API
			name := args[0]
			r, err := cli.Client.FetchFilter(name)
			if err != nil {
				return err
			}

			// Determine the format to use to output the data
			flag := helpers.GetChangedStringValueViper("format", cmd.Flags())
			format := cli.Config.Format()
			return helpers.PrintFormatted(flag, format, r, cmd.OutOrStdout(), printToList)
		},
	}

	helpers.AddFormatFlag(cmd.Flags())

	return cmd
}

func printToList(v interface{}, writer io.Writer) error {
	filter, ok := v.(*corev2.EventFilter)
	if !ok {
		return fmt.Errorf("%t is not an EventFilter", v)
	}
	cfg := &list.Config{
		Title: filter.Name,
		Rows: []*list.Row{
			{
				Label: "Name",
				Value: filter.Name,
			},
			{
				Label: "Namespace",
				Value: filter.Namespace,
			},
			{
				Label: "Action",
				Value: filter.Action,
			},
			{
				Label: "Expressions",
				Value: strings.Join(filter.Expressions, " && "),
			},
			{
				Label: "RuntimeAssets",
				Value: strings.Join(filter.RuntimeAssets, ", "),
			},
		},
	}

	return list.Print(writer, cfg)
}

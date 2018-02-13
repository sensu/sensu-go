package handler

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/list"
	"github.com/sensu/sensu-go/cli/elements/table"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// InfoCommand defines new check info command
func InfoCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "info [ID]",
		Short:        "show detailed handler information",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			// Fetch handler from API
			name := args[0]
			r, err := cli.Client.FetchHandler(name)
			if err != nil {
				return err
			}

			// Determine the format to use to output the data
			var format string
			if format = helpers.GetChangedStringValueFlag("format", cmd.Flags()); format == "" {
				format = cli.Config.Format()
			}

			if format == "json" {
				if err := helpers.PrintJSON(r, cmd.OutOrStdout()); err != nil {
					return err
				}
			} else {
				printCheckToList(r, cmd.OutOrStdout())
			}

			return nil
		},
	}

	helpers.AddFormatFlag(cmd.Flags())

	return cmd
}

func printCheckToList(handler *types.Handler, writer io.Writer) {
	// Determine what will be executed based on the type
	var execute string
	switch handler.Type {
	case types.HandlerTCPType:
		fallthrough
	case types.HandlerUDPType:
		execute = fmt.Sprintf(
			"%s %s://%s:%d",
			table.TitleStyle("PUSH:"),
			handler.Type,
			handler.Socket.Host,
			handler.Socket.Port,
		)
	case types.HandlerPipeType:
		execute = fmt.Sprintf(
			"%s  %s",
			table.TitleStyle("RUN:"),
			handler.Command,
		)
	case types.HandlerSetType:
		execute = fmt.Sprintf(
			"%s %s",
			table.TitleStyle("CALL:"),
			strings.Join(handler.Handlers, ","),
		)
	default:
		execute = "UNKNOWN"
	}

	cfg := &list.Config{
		Title: handler.Name,
		Rows: []*list.Row{
			{
				Label: "Name",
				Value: handler.Name,
			},
			{
				Label: "Type",
				Value: handler.Type,
			},
			{
				Label: "Timeout",
				Value: strconv.FormatInt(int64(handler.Timeout), 10),
			},
			{
				Label: "Filters",
				Value: strings.Join(handler.Filters, ", "),
			},
			{
				Label: "Mutator",
				Value: handler.Mutator,
			},
			{
				Label: "Execute",
				Value: execute,
			},
			{
				Label: "Environment Variables",
				Value: strings.Join(handler.EnvVars, ", "),
			},
		},
	}

	list.Print(writer, cfg)
}

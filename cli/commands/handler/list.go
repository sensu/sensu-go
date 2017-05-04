package handler

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/table"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// ListCommand defines new list events command
func ListCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "list",
		Short:        "list handlers",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			format, _ := cmd.Flags().GetString("format")

			// Fetch handlers from API
			r, err := cli.Client.ListHandlers()
			if err != nil {
				return err
			}

			if format == "json" {
				helpers.PrintJSON(r, cmd.OutOrStdout())
			} else {
				printHandlersToTable(r, cmd.OutOrStdout())
			}

			return nil
		},
	}

	helpers.AddFormatFlag(cmd.Flags(), cli.Config)

	return cmd
}

func printHandlersToTable(queryResults []types.Handler, writer io.Writer) {
	rows := make([]*table.Row, len(queryResults))
	for i, result := range queryResults {
		rows[i] = &table.Row{Value: result}
	}

	table := table.New([]*table.Column{
		{
			Title:       "Name",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				handler, _ := data.(types.Handler)
				return handler.Name
			},
		},
		{
			Title: "Type",
			CellTransformer: func(data interface{}) string {
				handler, _ := data.(types.Handler)
				return handler.Type
			},
		},
		{
			Title: "Timeout",
			CellTransformer: func(data interface{}) string {
				handler, _ := data.(types.Handler)
				return strconv.Itoa(handler.Timeout)
			},
		},
		{
			Title: "Mutator",
			CellTransformer: func(data interface{}) string {
				handler, _ := data.(types.Handler)
				return handler.Mutator
			},
		},
		{
			Title: "Execute",
			CellTransformer: func(data interface{}) string {
				handler, _ := data.(types.Handler)

				switch handler.Type {
				case "tcp":
					fallthrough
				case "udp":
					return fmt.Sprintf(
						"%s %s://%s:%d",
						table.TitleStyle("PUSH:"),
						handler.Type,
						handler.Socket.Host,
						handler.Socket.Port,
					)
				case "pipe":
					return fmt.Sprintf(
						"%s  %s",
						table.TitleStyle("RUN:"),
						handler.Command,
					)
				case "set":
					return fmt.Sprintf(
						"%s %s",
						table.TitleStyle("CALL:"),
						strings.Join(handler.Handlers, ","),
					)
				default:
					return "UNKNOWN"
				}
			},
		},
	})

	table.Render(writer, rows)
}

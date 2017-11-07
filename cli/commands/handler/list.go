package handler

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/flags"
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
			org := cli.Config.Organization()
			if ok, _ := cmd.Flags().GetBool(flags.AllOrgs); ok {
				org = "*"
			}

			// Fetch handlers from API
			r, err := cli.Client.ListHandlers(org)
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
				printHandlersToTable(r, cmd.OutOrStdout())
			}

			return nil
		},
	}

	helpers.AddFormatFlag(cmd.Flags())
	cmd.Flags().Bool(flags.AllOrgs, false, "Include records from all organizations")

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
				return strconv.FormatUint(uint64(handler.Timeout), 10)
			},
		},
		{
			Title: "Filters",
			CellTransformer: func(data interface{}) string {
				handler, _ := data.(types.Handler)
				return strings.Join(handler.Filters, ",")
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

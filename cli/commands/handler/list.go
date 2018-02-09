package handler

import (
	"errors"
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
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}
			org := cli.Config.Organization()
			if ok, _ := cmd.Flags().GetBool(flags.AllOrgs); ok {
				org = "*"
			}

			// Fetch handlers from API
			results, err := cli.Client.ListHandlers(org)
			if err != nil {
				return err
			}

			// Print the results based on the user preferences
			return helpers.Print(cmd, cli.Config.Format(), printToTable, results)
		},
	}

	helpers.AddFormatFlag(cmd.Flags())
	helpers.AddAllOrganization(cmd.Flags())

	return cmd
}

func printToTable(results interface{}, writer io.Writer) {
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
				case types.HandlerTCPType:
					fallthrough
				case types.HandlerUDPType:
					return fmt.Sprintf(
						"%s %s://%s:%d",
						table.TitleStyle("PUSH:"),
						handler.Type,
						handler.Socket.Host,
						handler.Socket.Port,
					)
				case types.HandlerPipeType:
					return fmt.Sprintf(
						"%s  %s",
						table.TitleStyle("RUN:"),
						handler.Command,
					)
				case types.HandlerSetType:
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

	table.Render(writer, results)
}

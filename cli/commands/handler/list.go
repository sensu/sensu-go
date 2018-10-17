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
			namespace := cli.Config.Namespace()
			if ok, _ := cmd.Flags().GetBool(flags.AllNamespaces); ok {
				namespace = types.NamespaceTypeAll
			}

			// Fetch handlers from API
			results, err := cli.Client.ListHandlers(namespace)
			if err != nil {
				return err
			}

			// Print the results based on the user preferences
			resources := []types.Resource{}
			for i := range results {
				resources = append(resources, &results[i])
			}
			return helpers.Print(cmd, cli.Config.Format(), printToTable, resources, results)
		},
	}

	helpers.AddFormatFlag(cmd.Flags())
	helpers.AddAllNamespace(cmd.Flags())

	return cmd
}

func printToTable(results interface{}, writer io.Writer) {
	table := table.New([]*table.Column{
		{
			Title:       "Name",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				handler, ok := data.(types.Handler)
				if !ok {
					return cli.TypeError
				}
				return handler.Name
			},
		},
		{
			Title: "Type",
			CellTransformer: func(data interface{}) string {
				handler, ok := data.(types.Handler)
				if !ok {
					return cli.TypeError
				}
				return handler.Type
			},
		},
		{
			Title: "Timeout",
			CellTransformer: func(data interface{}) string {
				handler, ok := data.(types.Handler)
				if !ok {
					return cli.TypeError
				}
				return strconv.FormatUint(uint64(handler.Timeout), 10)
			},
		},
		{
			Title: "Filters",
			CellTransformer: func(data interface{}) string {
				handler, ok := data.(types.Handler)
				if !ok {
					return cli.TypeError
				}
				return strings.Join(handler.Filters, ",")
			},
		},
		{
			Title: "Mutator",
			CellTransformer: func(data interface{}) string {
				handler, ok := data.(types.Handler)
				if !ok {
					return cli.TypeError
				}
				return handler.Mutator
			},
		},
		{
			Title: "Execute",
			CellTransformer: func(data interface{}) string {
				handler, ok := data.(types.Handler)
				if !ok {
					return cli.TypeError
				}
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
		{
			Title:       "Environment Variables",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				handler, ok := data.(types.Handler)
				if !ok {
					return cli.TypeError
				}
				return strings.Join(handler.EnvVars, ",")
			},
		},
	})

	table.Render(writer, results)
}

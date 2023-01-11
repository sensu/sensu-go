package handler

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/cli/commands/flags"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/table"

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
				namespace = corev2.NamespaceTypeAll
			}

			opts, err := helpers.ListOptionsFromFlags(cmd.Flags())
			if err != nil {
				return err
			}

			// Fetch handlers from API
			var header http.Header
			results := []corev2.Handler{}
			err = cli.Client.List(client.HandlersPath(namespace), &results, &opts, &header)
			if err != nil {
				return err
			}

			// Print the results based on the user preferences
			resources := []corev3.Resource{}
			for i := range results {
				resources = append(resources, &results[i])
			}
			return helpers.PrintList(cmd, cli.Config.Format(), printToTable, resources, results, header)
		},
	}

	helpers.AddFormatFlag(cmd.Flags())
	helpers.AddAllNamespace(cmd.Flags())
	helpers.AddFieldSelectorFlag(cmd.Flags())
	helpers.AddLabelSelectorFlag(cmd.Flags())
	helpers.AddChunkSizeFlag(cmd.Flags())

	return cmd
}

func printToTable(results interface{}, writer io.Writer) {
	table := table.New([]*table.Column{
		{
			Title:       "Name",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				handler, ok := data.(corev2.Handler)
				if !ok {
					return cli.TypeError
				}
				return handler.Name
			},
		},
		{
			Title: "Type",
			CellTransformer: func(data interface{}) string {
				handler, ok := data.(corev2.Handler)
				if !ok {
					return cli.TypeError
				}
				return handler.Type
			},
		},
		{
			Title: "Timeout",
			CellTransformer: func(data interface{}) string {
				handler, ok := data.(corev2.Handler)
				if !ok {
					return cli.TypeError
				}
				return strconv.FormatUint(uint64(handler.Timeout), 10)
			},
		},
		{
			Title: "Filters",
			CellTransformer: func(data interface{}) string {
				handler, ok := data.(corev2.Handler)
				if !ok {
					return cli.TypeError
				}
				return strings.Join(handler.Filters, ",")
			},
		},
		{
			Title: "Mutator",
			CellTransformer: func(data interface{}) string {
				handler, ok := data.(corev2.Handler)
				if !ok {
					return cli.TypeError
				}
				return handler.Mutator
			},
		},
		{
			Title: "Execute",
			CellTransformer: func(data interface{}) string {
				handler, ok := data.(corev2.Handler)
				if !ok {
					return cli.TypeError
				}
				switch handler.Type {
				case corev2.HandlerTCPType:
					fallthrough
				case corev2.HandlerUDPType:
					return fmt.Sprintf(
						"%s %s://%s:%d",
						table.TitleStyle("PUSH:"),
						handler.Type,
						handler.Socket.Host,
						handler.Socket.Port,
					)
				case corev2.HandlerPipeType:
					return fmt.Sprintf(
						"%s  %s",
						table.TitleStyle("RUN:"),
						handler.Command,
					)
				case corev2.HandlerSetType:
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
				handler, ok := data.(corev2.Handler)
				if !ok {
					return cli.TypeError
				}
				return strings.Join(handler.EnvVars, ",")
			},
		},
		{
			Title: "Assets",
			CellTransformer: func(data interface{}) string {
				handler, ok := data.(corev2.Handler)
				if !ok {
					return cli.TypeError
				}
				return strings.Join(handler.RuntimeAssets, ",")
			},
		},
	})

	table.Render(writer, results)
}

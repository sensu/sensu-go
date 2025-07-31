package logging

import (
	"errors"
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/table"
	"github.com/sensu/sensu-go/util/logging"
	"github.com/spf13/cobra"
	"io"
	"net/http"
)

// ListLogLevels sets log level for an module
func ListLogLevels(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "list",
		Short:        "list all modules and respective log levels",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			opts, err := helpers.ListOptionsFromFlags(cmd.Flags())
			if err != nil {
				return err
			}

			// Fetch log levels from API
			var header http.Header
			var results []logging.LogLevelRequest
			err = cli.Client.List(client.LogLevelPath("modules"), &results, &opts, &header)
			if err != nil {
				return err
			}

			// Print the results based on the user preferences
			var resources []corev2.Resource
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
			Title:       "Module",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				obj, ok := data.(logging.LogLevelRequest)
				if !ok {
					return cli.TypeError
				}
				return obj.Module
			},
		},
		{
			Title: "Log Level",
			CellTransformer: func(data interface{}) string {
				obj, ok := data.(logging.LogLevelRequest)
				if !ok {
					return cli.TypeError
				}
				return obj.Level
			},
		},
	})

	table.Render(writer, results)
}

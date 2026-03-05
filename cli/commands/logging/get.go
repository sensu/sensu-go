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

// GetModuleLogLevel gets log level for an module
func GetModuleLogLevel(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "get [command]",
		Short:        "get an module log level",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			// Fetch handlers from API
			var header http.Header
			var result logging.LogLevelRequest
			var results []logging.LogLevelRequest
			err := cli.Client.Get(client.LogLevelPath("modules", args[0]), &result)
			if err != nil {
				return err
			}

			// convert to array to reuse helper function PrintList
			results = append(results, result)

			// Print the results based on the user preferences
			var resources []corev2.Resource
			for i := range results {
				resources = append(resources, &results[i])
			}
			return helpers.PrintList(cmd, cli.Config.Format(), displayResult, resources, results, header)
		},
	}

	helpers.AddFormatFlag(cmd.Flags())
	helpers.AddAllNamespace(cmd.Flags())
	helpers.AddFieldSelectorFlag(cmd.Flags())
	helpers.AddLabelSelectorFlag(cmd.Flags())
	helpers.AddChunkSizeFlag(cmd.Flags())

	return cmd
}

func displayResult(results interface{}, writer io.Writer) {
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

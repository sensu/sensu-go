package fallbackPipeline

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/cli/commands/flags"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/table"

	"github.com/spf13/cobra"
)

// ListCommand defines new list pipelines command
func ListCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "list",
		Short:        "list fallbackPipelines",
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
				fmt.Println("ERROR encountered here======")
				return err
			}

			// Fetch pipelines from API
			var header http.Header
			results := []corev2.FallbackPipeline{}
			err = cli.Client.List(client.FallbackPipelinesPath(namespace), &results, &opts, &header)
			if err != nil {
				fmt.Println("ERROR encountered here 2======")
				return err
			}

			// Print the results based on the user preferences
			resources := []corev2.Resource{}
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
				fallbackPipeline, ok := data.(corev2.FallbackPipeline)
				if !ok {
					return cli.TypeError
				}
				return fallbackPipeline.GetName()
			},
		},
		{
			Title: "FallbackPipelineList",
			CellTransformer: func(data interface{}) string {
				fallbackPipeline, ok := data.(corev2.FallbackPipeline)
				if !ok {
					return cli.TypeError
				}
				plist := []string{}
				for _, pipelinelist := range fallbackPipeline.Pipelist {
					plist = append(plist, pipelinelist.Name)
				}
				return strings.Join(plist, ",")
			},
		},
	})

	table.Render(writer, results)
}

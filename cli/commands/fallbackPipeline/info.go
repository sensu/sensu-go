package fallbackPipeline

import (
	"errors"
	"fmt"
	"io"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/list"
	"github.com/spf13/cobra"
)

// InfoCommand defines new pipeline info command
func InfoCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "info [FALLBACKPIPELINE]",
		Short:        "show detailed fallback-pipeline information",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			// Fetch pipeline from API
			name := args[0]
			pipeline, err := cli.Client.FetchFallbackPipeline(name)
			if err != nil {
				return err
			}

			// Determine the format to use to output the data
			flag := helpers.GetChangedStringValueViper("format", cmd.Flags())
			format := cli.Config.Format()
			return helpers.PrintFormatted(flag, format, pipeline, cmd.OutOrStdout(), printToList)
		},
	}

	helpers.AddFormatFlag(cmd.Flags())

	return cmd
}

func printToList(v interface{}, writer io.Writer) error {
	pipeline, ok := v.(*corev2.FallbackPipeline)
	if !ok {
		return fmt.Errorf("%t is not a fallback-pipeline", v)
	}

	cfg := &list.Config{
		Title: pipeline.GetName(),
		Rows: []*list.Row{
			{
				Label: "Name",
				Value: pipeline.GetName(),
			},
			{
				Label: "Pipelines List",
				Value: "",
			},
		},
	}

	for _, pipelines := range pipeline.PipelineList {
		cfg.Rows = append(cfg.Rows, &list.Row{
			Label: fmt.Sprintf("  %s", pipelines.GetName()),
			Value: fmt.Sprintf("  %s ", pipelines.GetType()),
		})
	}

	return list.Print(writer, cfg)
}

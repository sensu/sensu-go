package pipeline

import (
	"errors"
	"fmt"
	"io"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/list"
	"github.com/spf13/cobra"
)

// InfoCommand defines new pipeline info command
func InfoCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "info [PIPELINE]",
		Short:        "show detailed pipeline information",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			// Fetch pipeline from API
			name := args[0]
			pipeline, err := cli.Client.FetchPipeline(name)
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
	pipeline, ok := v.(*corev2.Pipeline)
	if !ok {
		return fmt.Errorf("%t is not a Pipeline", v)
	}

	cfg := &list.Config{
		Title: pipeline.GetName(),
		Rows: []*list.Row{
			{
				Label: "Name",
				Value: pipeline.GetName(),
			},
			{
				Label: "Workflows",
				Value: "",
			},
		},
	}

	for _, workflow := range pipeline.GetWorkflows() {
		cfg.Rows = append(cfg.Rows, &list.Row{
			Label: fmt.Sprintf("  %s", workflow.GetName()),
			Value: "",
		})

		for _, filter := range workflow.Filters {
			cfg.Rows = append(cfg.Rows, &list.Row{
				Label: "    Filter",
				Value: filter.ResourceID(),
			})
		}

		mutator := ""
		if workflow.Mutator != nil {
			mutator = workflow.Mutator.ResourceID()
		}
		cfg.Rows = append(cfg.Rows, &list.Row{
			Label: "    Mutator",
			Value: mutator,
		})

		cfg.Rows = append(cfg.Rows, &list.Row{
			Label: "    Handler",
			Value: workflow.Handler.ResourceID(),
		})
	}

	return list.Print(writer, cfg)
}

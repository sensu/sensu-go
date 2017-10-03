package event

import (
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/list"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// ShowCommand defines new event info command
func ShowCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "info [ENTITY] [CHECK]",
		Short:        "show detailed event information",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			format, _ := cmd.Flags().GetString("format")

			if len(args) != 2 {
				cmd.Help()
				return fmt.Errorf("missing argument(s)")
			}

			// Fetch event from API
			entity := args[0]
			check := args[1]
			event, err := cli.Client.FetchEvent(entity, check)
			if err != nil {
				return err
			}

			if format == "json" {
				if err := helpers.PrintJSON(event, cmd.OutOrStdout()); err != nil {
					return err
				}
			} else {
				printEntityToList(event, cmd.OutOrStdout())
			}

			return nil
		},
	}

	helpers.AddFormatFlag(cmd.Flags(), cli.Config)

	return cmd
}

func printEntityToList(event *types.Event, writer io.Writer) {
	cfg := &list.Config{
		Title: fmt.Sprintf("%s - %s", event.Entity.ID, event.Check.Config.Name),
		Rows: []*list.Row{
			{
				Label: "Entity",
				Value: event.Entity.ID,
			},
			{
				Label: "Check",
				Value: event.Check.Config.Name,
			},
			{
				Label: "Output",
				Value: event.Check.Output,
			},
			{
				Label: "Status",
				Value: strconv.Itoa(event.Check.Status),
			},
			{
				Label: "Timestamp",
				Value: time.Unix(event.Timestamp, 0).String(),
			},
		},
	}

	list.Print(writer, cfg)
}

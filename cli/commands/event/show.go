package event

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
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
			if len(args) != 2 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			// Fetch event from API
			entity := args[0]
			check := args[1]
			event, err := cli.Client.FetchEvent(entity, check)
			if err != nil {
				return err
			}

			// Determine the format to use to output the data
			var format string
			if format = helpers.GetChangedStringValueFlag("format", cmd.Flags()); format == "" {
				format = cli.Config.Format()
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

	helpers.AddFormatFlag(cmd.Flags())

	return cmd
}

func printEntityToList(event *types.Event, writer io.Writer) {
	statusHistory := []string{}
	for _, entry := range event.Check.History {
		statusHistory = append(statusHistory, fmt.Sprint(entry.Status))
	}

	cfg := &list.Config{
		Title: fmt.Sprintf("%s - %s", event.Entity.ID, event.Check.Name),
		Rows: []*list.Row{
			{
				Label: "Entity",
				Value: event.Entity.ID,
			},
			{
				Label: "Check",
				Value: event.Check.Name,
			},
			{
				Label: "Output",
				Value: strings.TrimSuffix(event.Check.Output, "\n"),
			},
			{
				Label: "Status",
				Value: strconv.Itoa(int(event.Check.Status)),
			},
			{
				Label: "History",
				Value: strings.Join(statusHistory, ","),
			},
			{
				Label: "Silenced",
				Value: strconv.FormatBool(len(event.Silenced) > 0),
			},
			{
				Label: "Timestamp",
				Value: time.Unix(event.Timestamp, 0).String(),
			},
		},
	}

	if len(event.Silenced) > 0 {
		silencedBy := &list.Row{
			Label: "Silenced By",
			Value: strings.Join(event.Silenced, ", "),
		}
		cfg.Rows = append(cfg.Rows[:len(cfg.Rows)-1], silencedBy, cfg.Rows[len(cfg.Rows)-1])
	}

	list.Print(writer, cfg)
}

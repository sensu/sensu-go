package event

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	v2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/list"
	"github.com/spf13/cobra"
)

// InfoCommand defines new event info command
func InfoCommand(cli *cli.SensuCli) *cobra.Command {
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
			flag := helpers.GetChangedStringValueViper("format", cmd.Flags())
			format := cli.Config.Format()
			return helpers.PrintFormatted(flag, format, event, cmd.OutOrStdout(), printToList)
		},
	}

	helpers.AddFormatFlag(cmd.Flags())

	return cmd
}

func printToList(v interface{}, writer io.Writer) error {
	event, ok := v.(*v2.Event)
	if !ok {
		return fmt.Errorf("%t is not an Event", v)
	}
	statusHistory := []string{}
	for _, entry := range event.Check.History {
		statusHistory = append(statusHistory, fmt.Sprint(entry.Status))
	}

	cfg := &list.Config{
		Title: fmt.Sprintf("%s - %s", event.Entity.Name, event.Check.Name),
		Rows: []*list.Row{
			{
				Label: "Entity",
				Value: event.Entity.Name,
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
				Value: strconv.FormatBool(event.Check.IsSilenced),
			},
		},
	}

	if event.Check.IsSilenced {
		silencedBy := &list.Row{
			Label: "Silenced By",
			Value: strings.Join(event.Check.Silenced, ", "),
		}
		cfg.Rows = append(cfg.Rows, silencedBy)
	}

	var uuidVal string
	if id := event.GetUUID(); id != uuid.Nil {
		// Only populate the uuid if it's nonzero
		uuidVal = id.String()
	}
	cfg.Rows = append(cfg.Rows, []*list.Row{
		{
			Label: "Timestamp",
			Value: time.Unix(event.Timestamp, 0).String(),
		},
		{
			Label: "UUID",
			Value: uuidVal,
		}}...)

	return list.Print(writer, cfg)
}

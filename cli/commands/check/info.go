package check

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	v2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/globals"
	"github.com/sensu/sensu-go/cli/elements/list"
	"github.com/spf13/cobra"
)

// InfoCommand defines new check info command
func InfoCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "info [ID]",
		Short:        "show detailed check information",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			// Fetch handlers from API
			checkID := args[0]
			r, err := cli.Client.FetchCheck(checkID)
			if err != nil {
				return err
			}

			// Determine the format to use to output the data
			flag := helpers.GetChangedStringValueViper("format", cmd.Flags())
			format := cli.Config.Format()
			return helpers.PrintFormatted(flag, format, r, cmd.OutOrStdout(), printToList)
		},
	}

	helpers.AddFormatFlag(cmd.Flags())

	return cmd
}

func printToList(v interface{}, writer io.Writer) error {
	r, ok := v.(*v2.CheckConfig)
	if !ok {
		return fmt.Errorf("%t is not a CheckConfig", v)
	}
	cfg := &list.Config{
		Title: r.Name,
		Rows: []*list.Row{
			{
				Label: "Name",
				Value: r.Name,
			},
			{
				Label: "Interval",
				Value: strconv.FormatInt(int64(r.Interval), 10),
			},
			{
				Label: "Command",
				Value: r.Command,
			},
			{
				Label: "Cron",
				Value: r.Cron,
			},
			{
				Label: "Timeout",
				Value: strconv.FormatInt(int64(r.Timeout), 10),
			},
			{
				Label: "TTL",
				Value: strconv.FormatInt(int64(r.Ttl), 10),
			},
			{
				Label: "Subscriptions",
				Value: strings.Join(r.Subscriptions, ", "),
			},
			{
				Label: "Handlers",
				Value: strings.Join(r.Handlers, ", "),
			},
			{
				Label: "Runtime Assets",
				Value: strings.Join(r.RuntimeAssets, ", "),
			},
			{
				Label: "Hooks",
				Value: globals.FormatHookLists(r.CheckHooks),
			},
			{
				Label: "Publish?",
				Value: strconv.FormatBool(r.Publish),
			},
			{
				Label: "Stdin?",
				Value: strconv.FormatBool(r.Stdin),
			},
			{
				Label: "Proxy Entity Name",
				Value: r.ProxyEntityName,
			},
			{
				Label: "Namespace",
				Value: r.Namespace,
			},
			{
				Label: "Metric Format",
				Value: r.OutputMetricFormat,
			},
			{
				Label: "Metric Handlers",
				Value: strings.Join(r.OutputMetricHandlers, ", "),
			},
		},
	}

	return list.Print(writer, cfg)
}

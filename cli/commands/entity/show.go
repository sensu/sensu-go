package entity

import (
	"errors"
	"io"
	"strings"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/commands/timeutil"
	"github.com/sensu/sensu-go/cli/elements/globals"
	"github.com/sensu/sensu-go/cli/elements/list"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// ShowCommand defines new entity info command
func ShowCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "info [ID]",
		Short:        "show detailed entity information",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			// Fetch handlers from API
			entityID := args[0]
			r, err := cli.Client.FetchEntity(entityID)
			if err != nil {
				return err
			}

			// Determine the format to use to output the data
			var format string
			if format = helpers.GetChangedStringValueFlag("format", cmd.Flags()); format == "" {
				format = cli.Config.Format()
			}

			if format == "json" {
				if err := helpers.PrintJSON(r, cmd.OutOrStdout()); err != nil {
					return err
				}
			} else {
				printEntityToList(r, cmd.OutOrStdout())
			}

			return nil
		},
	}

	helpers.AddFormatFlag(cmd.Flags())

	return cmd
}

func printEntityToList(r *types.Entity, writer io.Writer) {
	cfg := &list.Config{
		Title: r.ID,
		Rows: []*list.Row{
			{
				Label: "ID",
				Value: r.ID,
			},
			{
				Label: "Class",
				Value: r.Class,
			},
			{
				Label: "Subscriptions",
				Value: strings.Join(r.Subscriptions, ", "),
			},
			{
				Label: "Last Seen",
				Value: timeutil.HumanTimestamp(r.LastSeen),
			},
			{
				Label: "Hostname",
				Value: r.System.Hostname,
			},
			{
				Label: "OS",
				Value: r.System.OS,
			},
			{
				Label: "Platform",
				Value: r.System.Platform,
			},
			{
				Label: "Platform Family",
				Value: r.System.PlatformFamily,
			},
			{
				Label: "Platform Version",
				Value: r.System.PlatformVersion,
			},
			// TODO: Network interfaces
			{
				Label: "Auto-Deregistration",
				Value: globals.BooleanStyleP(r.Deregister),
			},
			{
				Label: "Deregistration Handler",
				Value: r.Deregistration.Handler,
			},
		},
	}

	list.Print(writer, cfg)
}

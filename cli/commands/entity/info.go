package entity

import (
	"errors"
	"fmt"
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

// InfoCommand defines new entity info command
func InfoCommand(cli *cli.SensuCli) *cobra.Command {
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
			entityName := args[0]
			r, err := cli.Client.FetchEntity(entityName)
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
	r, ok := v.(*types.Entity)
	if !ok {
		return fmt.Errorf("%t is not an Entity", v)
	}
	cfg := &list.Config{
		Title: r.Name,
		Rows: []*list.Row{
			{
				Label: "Name",
				Value: r.Name,
			},
			{
				Label: "Entity Class",
				Value: r.EntityClass,
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

	return list.Print(writer, cfg)
}

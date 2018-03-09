package hook

import (
	"errors"
	"io"
	"strconv"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/globals"
	"github.com/sensu/sensu-go/cli/elements/list"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// InfoCommand defines new hook info command
func InfoCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "info [ID]",
		Short:        "show detailed hook information",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			// Fetch handlers from API
			hookID := args[0]
			r, err := cli.Client.FetchHook(hookID)
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
			}
			return printHookToList(r, cmd.OutOrStdout())
		},
	}

	helpers.AddFormatFlag(cmd.Flags())

	return cmd
}

func printHookToList(r *types.HookConfig, writer io.Writer) error {
	cfg := &list.Config{
		Title: r.Name,
		Rows: []*list.Row{
			{
				Label: "Name",
				Value: r.Name,
			},
			{
				Label: "Command",
				Value: r.Command,
			},
			{
				Label: "Timeout",
				Value: strconv.FormatInt(int64(r.Timeout), 10),
			},
			{
				Label: "Stdin?",
				Value: globals.BooleanStyleP(r.Stdin),
			},
			{
				Label: "Organization",
				Value: r.Organization,
			},
			{
				Label: "Environment",
				Value: r.Environment,
			},
		},
	}

	return list.Print(writer, cfg)
}

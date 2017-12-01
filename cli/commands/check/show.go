package check

import (
	"io"
	"strconv"
	"strings"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/list"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// ShowCommand defines new check info command
func ShowCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "info [ID]",
		Short:        "show detailed check information",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				cmd.Help()
				return nil
			}

			// Fetch handlers from API
			checkID := args[0]
			r, err := cli.Client.FetchCheck(checkID)
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
				printCheckToList(r, cmd.OutOrStdout())
			}

			return nil
		},
	}

	helpers.AddFormatFlag(cmd.Flags())

	return cmd
}

func printCheckToList(r *types.CheckConfig, writer io.Writer) {
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
				Label: "Publish?",
				Value: strconv.FormatBool(r.Publish),
			},
			{
				Label: "Stdin?",
				Value: strconv.FormatBool(r.Stdin),
			},
			{
				Label: "Source",
				Value: r.Source,
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

	list.Print(writer, cfg)
}

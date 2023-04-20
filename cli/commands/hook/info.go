package hook

import (
	"errors"
	"fmt"
	"io"
	"strconv"

	v2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/globals"
	"github.com/sensu/sensu-go/cli/elements/list"
	"github.com/spf13/cobra"
)

// InfoCommand defines new hook info command
func InfoCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:		"info [ID]",
		Short:		"show detailed hook information",
		SilenceUsage:	true,
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
			flag := helpers.GetChangedStringValueViper("format", cmd.Flags())
			format := cli.Config.Format()
			return helpers.PrintFormatted(flag, format, r, cmd.OutOrStdout(), printToList)
		},
	}

	helpers.AddFormatFlag(cmd.Flags())

	return cmd
}

func printToList(v interface{}, writer io.Writer) error {
	r, ok := v.(*v2.HookConfig)
	if !ok {
		return fmt.Errorf("%t is not a HookConfig", v)
	}
	cfg := &list.Config{
		Title:	r.Name,
		Rows: []*list.Row{
			{
				Label:	"Name",
				Value:	r.Name,
			},
			{
				Label:	"Command",
				Value:	r.Command,
			},
			{
				Label:	"Timeout",
				Value:	strconv.FormatInt(int64(r.Timeout), 10),
			},
			{
				Label:	"Stdin?",
				Value:	globals.BooleanStyleP(r.Stdin),
			},
			{
				Label:	"Namespace",
				Value:	r.Namespace,
			},
		},
	}

	return list.Print(writer, cfg)
}

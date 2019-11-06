package silenced

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/list"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// InfoCommand defines new silenced info command
func InfoCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "info [Name]",
		Short:        "show detailed silenced information",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			if len(args) == 0 {
				_ = cmd.Help()
				return errors.New("must provide silence name")
			}

			r, err := cli.Client.FetchSilenced(args[0])
			if err != nil {
				return err
			}

			// Determine the format to use to output the data
			flag := helpers.GetChangedStringValueFlag("format", cmd.Flags())
			format := cli.Config.Format()
			return helpers.PrintFormatted(flag, format, r, cmd.OutOrStdout(), printToList)
		},
	}

	helpers.AddFormatFlag(cmd.Flags())

	return cmd

}

func expireTime(beginTS, expireSeconds int64) time.Duration {
	begin := time.Unix(beginTS, 0)
	expire := time.Duration(expireSeconds) * time.Second
	if time.Now().Before(begin) {
		return (expire - time.Until(begin)).Truncate(time.Second)
	}
	return time.Duration(expireSeconds) * time.Second
}

func printToList(v interface{}, writer io.Writer) error {
	r, ok := v.(*types.Silenced)
	if !ok {
		return fmt.Errorf("%t is not a Silenced", v)
	}
	cfg := &list.Config{
		Title: r.Name,
		Rows: []*list.Row{
			{
				Label: "Expire",
				Value: expireTime(r.Begin, r.Expire).String(),
			},
			{
				Label: "ExpireOnResolve",
				Value: fmt.Sprintf("%t", r.ExpireOnResolve),
			},
			{
				Label: "Creator",
				Value: r.Creator,
			},
			{
				Label: "Check",
				Value: r.Check,
			},
			{
				Label: "Reason",
				Value: r.Reason,
			},
			{
				Label: "Subscription",
				Value: r.Subscription,
			},
			{
				Label: "Namespace",
				Value: r.Namespace,
			},
		},
	}
	if time.Now().Before(time.Unix(r.Begin, 0)) {
		extraRows := []*list.Row{{
			Label: "Begin",
			Value: time.Unix(r.Begin, 0).Format(time.RFC822),
		}}
		cfg.Rows = append(extraRows, cfg.Rows...)
	}

	return list.Print(writer, cfg)
}

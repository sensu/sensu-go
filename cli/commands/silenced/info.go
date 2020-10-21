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

			name, err := getName(cmd, args)
			if err != nil {
				return err
			}
			r, err := cli.Client.FetchSilenced(name)
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
	cmd.Flags().StringP("subscription", "s", "*", "name of the silenced subscription")
	cmd.Flags().StringP("check", "c", "*", "name of the silenced check")

	return cmd

}

func expireTime(beginTS, expireSeconds int64) string {
	// If we have no expiration, return -1
	if expireSeconds == -1 {
		return "-1"
	}

	begin := time.Unix(beginTS, 0)
	if time.Now().Before(begin) {
		// If the silenced entry is not yet in effect, because the being timestamp
		// is in the future, display the full expiration date as RFC3339
		expire := begin.Add(time.Duration(expireSeconds) * time.Second)
		return expire.Format(timeFormat)
	}

	// If the silenced entry is in effect, display its configured duration
	return (time.Duration(expireSeconds) * time.Second).String()
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
				Value: expireTime(r.Begin, r.Expire),
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
			Value: time.Unix(r.Begin, 0).Format(timeFormat),
		}}
		cfg.Rows = append(extraRows, cfg.Rows...)
	}

	return list.Print(writer, cfg)
}

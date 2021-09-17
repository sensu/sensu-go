package apikey

import (
	"errors"
	"fmt"
	"io"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/list"
	"github.com/spf13/cobra"
)

// InfoCommand adds a command that displays apikeys.
func InfoCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "info [NAME]",
		Short:        "show detailed api-key information",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			apikey := &corev2.APIKey{
				ObjectMeta: corev2.ObjectMeta{
					Name: args[0],
				},
			}
			err := cli.Client.Get(apikey.URIPath(), apikey)
			if err != nil {
				return err
			}

			flag := helpers.GetChangedStringValueViper("format", cmd.Flags())
			format := cli.Config.Format()
			return helpers.PrintFormatted(flag, format, apikey, cmd.OutOrStdout(), printToList)
		},
	}

	helpers.AddFormatFlag(cmd.Flags())

	return cmd
}

func printToList(v interface{}, writer io.Writer) error {
	r, ok := v.(*corev2.APIKey)
	if !ok {
		return fmt.Errorf("%t is not an APIKey", v)
	}
	cfg := &list.Config{
		Title: r.Name,
		Rows: []*list.Row{
			{
				Label: "Name",
				Value: r.Name,
			},
			{
				Label: "Username",
				Value: r.Username,
			},
			{
				Label: "Created At",
				Value: time.Unix(r.CreatedAt, 0).String(),
			},
		},
	}

	return list.Print(writer, cfg)
}

package tessen

import (
	"errors"
	"fmt"
	"io"
	"strconv"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/list"
	"github.com/spf13/cobra"
)

// InfoCommand provides the tessen configuration
func InfoCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "info",
		Short:        "show tessen configuration",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			config := &corev2.TessenConfig{}
			if err := cli.Client.Get(config.URIPath(), config); err != nil {
				return err
			}

			// Determine the format to use to output the data
			flag := helpers.GetChangedStringValueViper("format", cmd.Flags())
			format := cli.Config.Format()
			return helpers.PrintFormatted(flag, format, config, cmd.OutOrStdout(), printToList)
		},
	}

	helpers.AddFormatFlag(cmd.Flags())

	return cmd
}

func printToList(v interface{}, writer io.Writer) error {
	r, ok := v.(*corev2.TessenConfig)
	if !ok {
		return fmt.Errorf("%t is not a tessen config", v)
	}
	cfg := &list.Config{
		Title: "Tessen Configuration",
		Rows: []*list.Row{
			{
				Label: "Opt-Out",
				Value: strconv.FormatBool(r.OptOut),
			},
		},
	}

	return list.Print(writer, cfg)
}

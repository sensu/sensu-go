package opampagent

import (
	"errors"
	"fmt"
	"io"

	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/list"
	"github.com/spf13/cobra"
)

func ListConfigCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "config",
		Short:        "show opamp agent config",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				cmd.Help()
				return errors.New("invalid argument(s)")
			}

			config := &corev3.OpampAgentConfig{}
			if err := cli.Client.Get(config.URIPath(), config); err != nil {
				return err
			}

			flag := helpers.GetChangedStringValueViper("format", cmd.Flags())
			format := cli.Config.Format()
			return helpers.PrintFormatted(flag, format, config, cmd.OutOrStdout(), print)
		},
	}

	helpers.AddFormatFlag(cmd.Flags())
	return cmd
}

func print(v interface{}, w io.Writer) error {
	cfg, ok := v.(*corev3.OpampAgentConfig)
	if !ok {
		return fmt.Errorf("unexpected type %t. expected opamp agent config", v)
	}
	return list.Print(w, &list.Config{
		Title: "Opamp Agent Configuration",
		Rows: []*list.Row{
			{
				Label: "Body",
				Value: cfg.Body,
			},
			{
				Label: "Content-Type",
				Value: cfg.ContentType,
			},
		},
	})
}

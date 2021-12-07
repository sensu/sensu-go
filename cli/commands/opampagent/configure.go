package opampagent

import (
	"os"

	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

func ConfigureCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "configure -file OpampAgentConfig.yaml",
		Short:        "update the global OTEL Collector configuration file",
		SilenceUsage: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			_ = cmd.MarkFlagRequired("file")
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			fPath, err := cmd.Flags().GetString("file")
			if err != nil {
				return err
			}
			contentType, err := cmd.Flags().GetString("content-type")
			if err != nil {
				return err
			}

			contents, err := os.ReadFile(fPath)
			if err != nil {
				return err
			}
			config := &corev3.OpampAgentConfig{
				Body:        string(contents),
				ContentType: contentType,
			}
			if err := cli.Client.Put(config.URIPath(), config); err != nil {
				return err
			}

			return nil
		},
	}
	_ = cmd.Flags().StringP("file", "f", "", "otel collector configuration file")
	_ = cmd.Flags().StringP("content-type", "c", "yaml", "otel collector configuration file content type")

	return cmd
}

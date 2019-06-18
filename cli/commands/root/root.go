package root

import (
	colorable "github.com/mattn/go-colorable"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client/config"
	"github.com/sensu/sensu-go/cli/commands/version"
	"github.com/sensu/sensu-go/util/path"
	"github.com/spf13/cobra"
)

// Command defines the root command for sensuctl
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:          cli.SensuCmdName,
		Short:        cli.SensuCmdName + " controls Sensu instances",
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	// Use colorable (fixes Windows color support)
	cmd.SetOut(colorable.NewColorableStdout())
	cmd.SetErr(colorable.NewColorableStderr())

	// Templates
	cmd.SetUsageTemplate(usageTemplate)

	// Version command
	cmd.AddCommand(version.Command())

	// Global flags
	cmd.PersistentFlags().String("api-url", "", "host URL of Sensu installation")
	cmd.PersistentFlags().String("trusted-ca-file", "", "TLS CA certificate bundle in PEM format")
	cmd.PersistentFlags().Bool("insecure-skip-tls-verify", false, "skip TLS certificate verification (not recommended!)")
	cmd.PersistentFlags().String("config-dir", path.UserConfigDir("sensuctl"), "path to directory containing configuration files")
	cmd.PersistentFlags().String("cache-dir", path.UserCacheDir("sensuctl"), "path to directory containing cache & temporary files")
	cmd.PersistentFlags().String("namespace", config.DefaultNamespace, "namespace in which we perform actions")

	return cmd
}

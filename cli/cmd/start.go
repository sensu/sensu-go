package main

import (
	"fmt"
	"os"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client/config"
	"github.com/sensu/sensu-go/cli/commands"
	hooks "github.com/sensu/sensu-go/cli/commands/hooks"
	"github.com/sensu/sensu-go/util/path"
	"github.com/sensu/sensu-go/version"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := configureRootCmd()
	sensuCli := cli.New(rootCmd.PersistentFlags())

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		return hooks.ConfigurationPresent(cmd, sensuCli)
	}

	commands.AddCommands(rootCmd, sensuCli)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func configureRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          cli.SensuCmdName,
		Short:        cli.SensuCmdName + " controls Sensu instances",
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	// Templates
	cmd.SetUsageTemplate(usageTemplate)

	// Version command
	cmd.AddCommand(newVersionCommand())

	// Global flags
	cmd.PersistentFlags().String("api-url", "", "host URL of Sensu installation")
	cmd.PersistentFlags().String("config-dir", path.UserConfigDir("sensuctl"), "path to directory containing configuration files")
	cmd.PersistentFlags().String("cache-dir", path.UserCacheDir("sensuctl"), "path to directory containing cache & temporary files")
	cmd.PersistentFlags().String("organization", config.DefaultOrganization, "organization in which we perform actions")
	cmd.PersistentFlags().String("environment", config.DefaultEnvironment, "environment in which we perform actions")

	return cmd
}

func newVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show the sensu-ctl version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("sensu-ctl version %s, build %s, built %s\n",
				version.Semver(),
				version.BuildSHA,
				version.BuildDate,
			)
		},
	}

	return cmd
}

var usageTemplate = `Usage:

{{- if not .HasSubCommands}}	{{.UseLine}}{{end}}
{{- if .HasSubCommands}}	{{ .CommandPath}} COMMAND{{end}}

{{ .Short | trim }}

{{- if .HasFlags}}

Options:
{{ wrappedFlagUsages . | trimRightSpace}}

{{- end}}

{{- if hasOperationalSubCommands . }}

Commands:

{{- range operationalSubCommands . }}
  {{rpad .Name .NamePadding }} {{.Short}}
{{- end}}
{{- end}}

{{- if hasManagementSubCommands . }}

Management Commands:

{{- range managementSubCommands . }}
  {{rpad .Name .NamePadding }} {{.Short}}
{{- end}}
{{- end}}

{{- if .HasSubCommands }}

Run '{{.CommandPath}} COMMAND --help' for more information on a command.
{{- end}}
`

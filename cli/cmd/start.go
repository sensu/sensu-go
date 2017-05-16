package main

import (
	"fmt"
	"os"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands"
	hooks "github.com/sensu/sensu-go/cli/commands/hooks"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := configureRootCmd()
	sensuCli := cli.New(rootCmd.PersistentFlags())

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		if err := hooks.ConfigurationPresent(cmd, sensuCli); err != nil {
			return err
		}

		return nil
	}

	commands.AddCommands(rootCmd, sensuCli)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func configureRootCmd() *cobra.Command {
	showVersion := false
	cmd := &cobra.Command{
		Use:          "sensu-cli",
		Short:        "A tool to help manage Sensu",
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			if showVersion {
				showCLIVersion()
			} else {
				cmd.Help()
			}
		},
	}

	// Templates
	cmd.SetUsageTemplate(usageTemplate)

	// Version flag
	cmd.Flags().BoolVarP(&showVersion, "version", "v", false, "print version information")

	// Global flags
	cmd.PersistentFlags().StringP("api-url", "", "", "host URL of Sensu installation")
	cmd.PersistentFlags().StringP("api-secret", "", "", "secret used to authorize your requests to your specified Sensu installation")
	cmd.PersistentFlags().StringP("profile", "", "default", "configuration values to use")

	return cmd
}

func showCLIVersion() {
	// TODO: 😰
	fmt.Printf("Sensu CLI version %s\n", "0.1.alpha")
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

Managment Commands:

{{- range managementSubCommands . }}
  {{rpad .Name .NamePadding }} {{.Short}}
{{- end}}
{{- end}}

{{- if .HasSubCommands }}

Run '{{.CommandPath}} COMMAND --help' for more information on a command.
{{- end}}
`

var helpTemplate = `
{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}`

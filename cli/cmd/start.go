package main

import (
	"fmt"
	"os"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := configureRootCmd()
	sensuCli := cli.New(rootCmd.PersistentFlags())

	commands.AddCommands(rootCmd, sensuCli)

	if err := rootCmd.Execute(); err != nil {
		logger.Fatal(err.Error())
		os.Exit(1)
	}
}

func configureRootCmd() *cobra.Command {
	showVersion := false
	cmd := &cobra.Command{
		Use:   "sensu-cli",
		Short: "A tool to help manage Sensu",
		Run: func(cmd *cobra.Command, args []string) {
			if showVersion {
				showCLIVersion()
			} else {
				cmd.HelpFunc()(cmd, args)
			}
		},
	}

	// Templates
	cmd.SetUsageTemplate(usageTemplate)

	// Version flag
	cmd.Flags().BoolVarP(&showVersion, "version", "v", false, "Print version information")

	// Global flags
	cmd.PersistentFlags().StringP("baseURL", "", "", "Host URL of Sensu installation")
	cmd.PersistentFlags().StringP("profile", "", "default", "Configuration values to use")

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

{{- if hasManagementSubCommands . }}

Managment Commands:

{{- range managementSubCommands . }}
  {{rpad .Name .NamePadding }} {{.Short}}
{{- end}}
{{- end}}

{{- if hasOperationalSubCommands . }}

Commands:

{{- range operationalSubCommands . }}
  {{rpad .Name .NamePadding }} {{.Short}}
{{- end}}
{{- end}}

{{- if .HasSubCommands }}

Run '{{.CommandPath}} COMMAND --help' for more information on a command.
{{- end}}
`

var helpTemplate = `
{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}`

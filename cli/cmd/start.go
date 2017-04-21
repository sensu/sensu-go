package main

import (
	"os"

	"github.com/sensu/sensu-go/cli/cmd/commands"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "sensu-cli",
		Short: "A tool to help manage Sensu",
	}

	rootCmd.SetUsageTemplate(usageTemplate)
	rootCmd.SetHelpTemplate(helpTemplate)

	commands.AddCommands(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		logger.Fatal(err.Error())
		os.Exit(1)
	}
}

var usageTemplate = `Usage:

{{- if not .HasSubCommands}}	{{.UseLine}}{{end}}
{{- if .HasSubCommands}}	{{ .CommandPath}} COMMAND{{end}}

{{ .Short | trim }}

{{- if .HasFlags}}

Options:
{{ wrappedFlagUsages . | trimRightSpace}}

{{- end}}

{{- if .HasSubCommands }}

Commands:

{{- range subCommands . }}
  {{rpad .Name .NamePadding }} {{.Short}}
{{- end}}
{{- end}}

{{- if .HasSubCommands }}

Run '{{.CommandPath}} COMMAND --help' for more information on a command.
{{- end}}
`

var helpTemplate = `
{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}`

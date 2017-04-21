package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "sensu-cli",
		Short: "A tool to help manage Sensu",
	}

	rootCmd.SetUsageTemplate(usageTemplate)
	rootCmd.SetHelpTemplate(helpTemplate)
	rootCmd.AddCommand(eventCommand())

	if err := rootCmd.Execute(); err != nil {
		logger.Fatal(err.Error())
		os.Exit(1)
	}
}

func eventCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "event",
		Short: "returns a list of event sub-commands",
	}

	cmd.AddCommand(listEventsCommand())

	return cmd
}

func listEventsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list events",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("LISTING THE EVENTS\n", args)
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

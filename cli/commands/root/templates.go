package root

import (
	"github.com/docker/docker/pkg/term"
	"github.com/spf13/cobra"
)

var usageTemplate = `Usage:

{{- if not .HasSubCommands}}	{{.UseLine}}{{end}}
{{- if .HasSubCommands}}	{{ .CommandPath}} COMMAND{{end}}

{{- if .HasAvailableLocalFlags}}

Flags:
{{ wrappedLocalFlagUsages . | trimRightSpace}}

{{- end}}

{{- if .HasAvailableInheritedFlags}}

Global Flags:
{{ wrappedInheritedFlagUsages . | trimRightSpace}}

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

{{- if hasEnterpriseSubCommands . }}

Enterprise Commands:

{{- range enterpriseSubCommands . }}
  {{rpad .Name .NamePadding }} {{.Short}}
{{- end}}
{{- end}}

{{- if .HasSubCommands }}

Run '{{.CommandPath}} COMMAND --help' for more information on a command.
{{- end}}

`

func init() {
	cobra.AddTemplateFunc("enterpriseSubCommands", enterpriseSubCommands)
	cobra.AddTemplateFunc("hasEnterpriseSubCommands", hasEnterpriseSubCommands)
	cobra.AddTemplateFunc("hasOperationalSubCommands", hasOperationalSubCommands)
	cobra.AddTemplateFunc("hasManagementSubCommands", hasManagementSubCommands)
	cobra.AddTemplateFunc("operationalSubCommands", operationalSubCommands)
	cobra.AddTemplateFunc("managementSubCommands", managementSubCommands)
	cobra.AddTemplateFunc("wrappedInheritedFlagUsages", wrappedInheritedFlagUsages)
	cobra.AddTemplateFunc("wrappedLocalFlagUsages", wrappedLocalFlagUsages)
}

func enterpriseSubCommands(cmd *cobra.Command) []*cobra.Command {
	return []*cobra.Command{}
}

func hasEnterpriseSubCommands(cmd *cobra.Command) bool {
	return false
}

func hasOperationalSubCommands(cmd *cobra.Command) bool {
	return len(operationalSubCommands(cmd)) > 0
}

func hasManagementSubCommands(cmd *cobra.Command) bool {
	return len(managementSubCommands(cmd)) > 0
}

func operationalSubCommands(cmd *cobra.Command) []*cobra.Command {
	cmds := []*cobra.Command{}
	for _, sub := range cmd.Commands() {
		if sub.IsAvailableCommand() && !sub.HasSubCommands() {
			cmds = append(cmds, sub)
		}
	}
	return cmds
}

func managementSubCommands(cmd *cobra.Command) []*cobra.Command {
	cmds := []*cobra.Command{}
	for _, sub := range cmd.Commands() {
		if sub.IsAvailableCommand() && sub.HasSubCommands() {
			cmds = append(cmds, sub)
		}
	}
	return cmds
}

func wrappedInheritedFlagUsages(cmd *cobra.Command) string {
	width := 80
	if ws, err := term.GetWinsize(0); err == nil {
		width = int(ws.Width)
	}
	return cmd.InheritedFlags().FlagUsagesWrapped(width - 1)
}

func wrappedLocalFlagUsages(cmd *cobra.Command) string {
	width := 80
	if ws, err := term.GetWinsize(0); err == nil {
		width = int(ws.Width)
	}
	return cmd.LocalFlags().FlagUsagesWrapped(width - 1)
}

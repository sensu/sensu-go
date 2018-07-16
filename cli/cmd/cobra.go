package main

import (
	"github.com/docker/docker/pkg/term"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

func init() {
	cobra.AddTemplateFunc("hasEnterpriseSubCommands", hasEnterpriseSubCommands)
	cobra.AddTemplateFunc("hasManagementSubCommands", hasManagementSubCommands)
	cobra.AddTemplateFunc("hasOperationalSubCommands", hasOperationalSubCommands)
	cobra.AddTemplateFunc("enterpriseSubCommands", enterpriseSubCommands)
	cobra.AddTemplateFunc("managementSubCommands", managementSubCommands)
	cobra.AddTemplateFunc("operationalSubCommands", operationalSubCommands)
	cobra.AddTemplateFunc("wrappedInheritedFlagUsages", wrappedInheritedFlagUsages)
	cobra.AddTemplateFunc("wrappedLocalFlagUsages", wrappedLocalFlagUsages)
}

func hasEnterpriseSubCommands(cmd *cobra.Command) bool {
	return len(enterpriseSubCommands(cmd)) > 0
}

func hasOperationalSubCommands(cmd *cobra.Command) bool {
	return len(operationalSubCommands(cmd)) > 0
}

func hasManagementSubCommands(cmd *cobra.Command) bool {
	return len(managementSubCommands(cmd)) > 0
}

func enterpriseSubCommands(cmd *cobra.Command) []*cobra.Command {
	cmds := []*cobra.Command{}
	for _, sub := range cmd.Commands() {
		if sub.IsAvailableCommand() && sub.HasSubCommands() {
			// Verify if the command has an annotation that contains the edition
			if edition, ok := sub.Annotations["edition"]; ok {
				// Only add commands explicitly marked as being enterprise
				if edition == types.EnterpriseEdition {
					cmds = append(cmds, sub)
				}
			}
		}
	}
	return cmds
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
			// Verify if the command has an annotation that contains the edition
			if edition, ok := sub.Annotations["edition"]; ok {
				// Ignore commands explicitly marked as being enterprise
				if edition == types.EnterpriseEdition {
					continue
				}
			}
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

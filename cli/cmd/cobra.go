package main

import (
	"github.com/docker/docker/pkg/term"
	"github.com/spf13/cobra"
)

func init() {
	cobra.AddTemplateFunc("subCommands", subCommands)
	cobra.AddTemplateFunc("wrappedFlagUsages", wrappedFlagUsages)
}

func subCommands(cmd *cobra.Command) []*cobra.Command {
	cmds := []*cobra.Command{}
	for _, sub := range cmd.Commands() {
		if sub.IsAvailableCommand() && sub.HasSubCommands() {
			cmds = append(cmds, sub)
		}
	}
	return cmds
}

func wrappedFlagUsages(cmd *cobra.Command) string {
	width := 80
	if ws, err := term.GetWinsize(0); err == nil {
		width = int(ws.Width)
	}
	return cmd.Flags().FlagUsagesWrapped(width - 1)
}

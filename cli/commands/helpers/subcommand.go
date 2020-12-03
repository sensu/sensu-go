package helpers

import (
	"fmt"

	"github.com/spf13/cobra"
)

func FindSuggestionsFor(cmd *cobra.Command, arg string) string {
	if cmd.DisableSuggestions {
		return ""
	}
	if cmd.SuggestionsMinimumDistance <= 0 {
		cmd.SuggestionsMinimumDistance = 2
	}
	suggestionsString := ""
	if suggestions := cmd.SuggestionsFor(arg); len(suggestions) > 0 {
		suggestionsString += "\n\nDid you mean this?\n"
		for _, s := range suggestions {
			suggestionsString += fmt.Sprintf("\t%v\n", s)
		}
	}
	return suggestionsString
}

func RunSubCommandE(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		cmd.Help()
		return nil
	}
	suggestions := FindSuggestionsFor(cmd, args[0])
	help := fmt.Sprintf("Run '%s --help' for usage.", cmd.CommandPath())
	return fmt.Errorf("unknown subcommand: %q for %q%s\n%s",
		args[0], cmd.CommandPath(), suggestions, help)
}

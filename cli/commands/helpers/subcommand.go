package helpers

import (
	"fmt"
	"os"
	"strings"

	"github.com/sensu/sensu-go/command"
	"github.com/spf13/cobra"
)

func UnknownCommandError(cmd *cobra.Command, args []string) error {
	cmdName := strings.Join(args, " ")
	cmdPath := cmd.CommandPath()
	suggestions := FindSuggestionsFor(cmd, string(args[0]))
	return &command.UsageError{
		Message: fmt.Sprintf(
			"unknown command %q for %q%s\nRun '%s --help' for usage.",
			cmdName, cmdPath, suggestions, cmdPath),
	}
}

// DefaultSubCommandRun prints the help message to stderr, hides error output,
// and then returns a command.UsageError when no arguments are given. If invalid
// arguments are given (e.g. anything other than "help") a command.UsageError
// will be returned with an "unknown command" message.
func DefaultSubCommandRunE(cmd *cobra.Command, args []string) error {
	// Handle no arguments
	if err := cobra.NoArgs(cmd, args); err == nil {
		cmd.SetOutput(os.Stderr)
		_ = cmd.Help()
		cmd.SilenceErrors = true
		return &command.UsageError{}
	}

	// Handle invalid arguments
	cmd.ValidArgs = []string{"help"}
	if err := cobra.OnlyValidArgs(cmd, args); err != nil {
		cmd.SetOutput(os.Stderr)
		return UnknownCommandError(cmd, args)
	}

	// Handle "help" argument
	_ = cmd.Help()
	return nil
}

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

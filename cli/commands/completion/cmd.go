package completion

import (
	"fmt"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/hooks"
	"github.com/spf13/cobra"
)

const (
	longUsage = `
Output shell completion code for the given shell. The code must be evaluated
to provide interactive completion of sensu-cli commands. This can be done by
sourcing it from the .bash_profile or .zshrc.

For help using with ZSH:

    $ ` + cli.SensuCmdName + ` completion zsh -h

For help using with Bash:

    $ ` + cli.SensuCmdName + ` completion bash -h
	`

	zshShell  = "zsh"
	bashShell = "bash"
)

// Command defines new command to help installing completions in shell
func Command(rootCmd *cobra.Command) *cobra.Command {
	exec := &completionExecutor{rootCmd: rootCmd}
	cmd := &cobra.Command{
		Use:   "completion",
		Short: "Output shell completion code for the specified shell (bash or zsh)",
		RunE:  exec.run,
		Annotations: map[string]string{
			// We want to be able to run this command regardless of whether the CLI
			// has been configured.
			hooks.ConfigurationRequirement: hooks.ConfigurationNotRequired,
		},
	}

	cmd.SetHelpFunc(exec.runHelp)

	return cmd
}

type completionExecutor struct {
	rootCmd *cobra.Command
}

func (e *completionExecutor) run(cmd *cobra.Command, args []string) error {
	shell, err := extractShell(args, 0)

	if shell == zshShell {
		return genZshCompletion(e.rootCmd)
	} else if shell == bashShell {
		return genBashCompletion(e.rootCmd)
	} else if err != nil {
		fmt.Fprintf(
			cmd.OutOrStderr(),
			"Error: %s\nRun '%v --help' for usage.\n'",
			err,
			cmd.CommandPath(),
		)
	} else {
		e.runHelp(cmd, args)
	}

	return nil
}

func (e *completionExecutor) runHelp(cmd *cobra.Command, args []string) {
	stdErr := cmd.OutOrStderr()

	if shell, err := extractShell(args, 1); err != nil {
		fmt.Fprintf(stdErr, "%s\n\n%s\n", err, longUsage)
	} else if shell == zshShell {
		fmt.Fprintln(stdErr, zshUsage)
	} else if shell == bashShell {
		fmt.Fprintln(stdErr, bashUsage)
	} else {
		fmt.Fprintln(stdErr, longUsage)
	}
}

func extractShell(args []string, i int) (string, error) {
	if len(args) > i {
		shell := args[i]
		if shell == zshShell || shell == bashShell {
			return shell, nil
		}
		return shell, fmt.Errorf("unknown shell: %q", shell)
	}

	return "", nil
}

package v2

import (
	"os"
	"strings"
)

// FakeHandlerCommand takes a command and (optionally) command args and will
// execute the TestHelperHandlerProcess test within the package
// FakeHandlerCommand is called from.
func FakeHandlerCommand(command string, args ...string) *Handler {
	cs := []string{os.Args[0], "-test.run=TestHelperHandlerProcess", "--", command}
	cs = append(cs, args...)
	cmdStr := strings.Join(cs, " ")
	trimmedCmd := strings.Trim(cmdStr, " ")
	env := []string{"GO_WANT_HELPER_HANDLER_PROCESS=1"}

	handler := &Handler{
		Command: trimmedCmd,
		EnvVars: env,
	}

	return handler
}

// FakeMutatorCommand takes a command and (optionally) command args and will
// execute the TestHelperMutatorProcess test within the package
// FakeMutatorCommand is called from.
func FakeMutatorCommand(command string, args ...string) *Mutator {
	cs := []string{os.Args[0], "-test.run=TestHelperMutatorProcess", "--", command}
	cs = append(cs, args...)
	cmdStr := strings.Join(cs, " ")
	trimmedCmd := strings.Trim(cmdStr, " ")
	env := []string{"GO_WANT_HELPER_MUTATOR_PROCESS=1"}

	mutator := &Mutator{
		Command: trimmedCmd,
		EnvVars: env,
	}

	return mutator
}

package command

import (
	"os"
	"strings"
)

// FakeCommand takes a command and (optionally) command args and will execute
// the TestHelperProcess test within the package FakeCommand is called from.
func FakeCommand(command string, args ...string) ExecutionRequest {
	cs := []string{os.Args[0], "-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmdStr := strings.Join(cs, " ")
	trimmedCmd := strings.Trim(cmdStr, " ")
	env := []string{"GO_WANT_HELPER_PROCESS=1"}

	execution := ExecutionRequest{
		Command: trimmedCmd,
		Env:     env,
	}

	return execution
}

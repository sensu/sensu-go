package completion

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommand(t *testing.T) {
	assert := assert.New(t)

	exCmd := &cobra.Command{}
	cmd := Command(exCmd)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("completion", cmd.Use)
	assert.Regexp("shell completion code", cmd.Short)
}

type executorTest struct {
	exec    *completionExecutor
	rootCmd *cobra.Command
	cmd     *cobra.Command
	out     *exWriter
}

func newExecutorTest() *executorTest {
	test := &executorTest{}
	test.rootCmd = &cobra.Command{
		Use:          "sensuctl",
		Short:        "sensuctl test test tests",
		SilenceUsage: true,
	}
	test.cmd = Command(test.rootCmd)
	test.exec = &completionExecutor{rootCmd: test.rootCmd}

	test.out = &exWriter{}
	test.cmd.SetOutput(test.out)
	test.rootCmd.SetOutput(test.out)

	return test
}

func TestRunWithNoArguments(t *testing.T) {
	test := newExecutorTest()
	err := test.exec.run(test.cmd, []string{})
	out := test.out.result

	require.NoError(t, err)
	require.NotEmpty(t, out)
	assert.Contains(t, out, "help")
}

func TestRunWithArgZsh(t *testing.T) {
	test := newExecutorTest()
	err := test.exec.run(test.cmd, []string{"zsh"})
	out := test.out.result

	require.NoError(t, err)
	require.NotEmpty(t, out)
	assert.Contains(t, out, "BASH_COMPLETION_EOF")
	assert.Contains(t, out, "convert_bash_to_zsh")
}

func TestRunWithArgBash(t *testing.T) {
	test := newExecutorTest()
	err := test.exec.run(test.cmd, []string{"bash"})
	out := test.out.result

	require.NoError(t, err)
	require.NotEmpty(t, out)
	assert.Contains(t, out, "_init_completion")
}

func TestRunWithBadArg(t *testing.T) {
	test := newExecutorTest()
	err := test.exec.run(test.cmd, []string{"fish"})
	out := test.out.result

	require.NoError(t, err)
	require.NotEmpty(t, out)
	assert.Contains(t, out, "unknown shell")
	assert.Contains(t, out, "usage")
}

func TestHelpWithNoArgs(t *testing.T) {
	test := newExecutorTest()
	test.exec.runHelp(test.cmd, []string{"completion"})
	out := test.out.result

	require.NotEmpty(t, out)
	assert.Contains(t, out, "Output shell completion code for the given shell")
	assert.Contains(t, out, "help")
}

func TestHelpWithZsh(t *testing.T) {
	test := newExecutorTest()
	test.exec.runHelp(test.cmd, []string{"completion", "zsh"})
	out := test.out.result

	require.NotEmpty(t, out)
	assert.Contains(t, out, "Add the following")
	assert.Contains(t, out, "zshrc")
}

func TestHelpWithBash(t *testing.T) {
	test := newExecutorTest()
	test.exec.runHelp(test.cmd, []string{"completion", "bash"})
	out := test.out.result

	require.NotEmpty(t, out)
	assert.Contains(t, out, "add the following")
	assert.Contains(t, out, "bash_profile")
}

func TestHelpWithBadArg(t *testing.T) {
	test := newExecutorTest()
	test.exec.runHelp(test.cmd, []string{"completion", "fish"})
	out := test.out.result

	require.NotEmpty(t, out)
	assert.Contains(t, out, "unknown shell")
	assert.Contains(t, out, "help")
}

type exWriter struct {
	result string
}

func (w *exWriter) Clean() {
	w.result = ""
}

func (w *exWriter) Write(p []byte) (int, error) {
	w.result += string(p)
	return len(w.result), nil
}

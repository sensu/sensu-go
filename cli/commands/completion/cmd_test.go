package completion

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
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

type ExecutorSuite struct {
	suite.Suite
	exec    *completionExecutor
	rootCmd *cobra.Command
	cmd     *cobra.Command
	out     *exWriter
}

func (suite *ExecutorSuite) SetupTest() {
	suite.rootCmd = &cobra.Command{}
	suite.cmd = Command(suite.rootCmd)
	suite.exec = &completionExecutor{rootCmd: suite.rootCmd}

	suite.out = &exWriter{}
	suite.cmd.SetOutput(suite.out)
	suite.rootCmd.SetOutput(suite.out)
}

func (suite *ExecutorSuite) TestRunWithNoArguments() {
	err := suite.exec.run(suite.cmd, []string{})
	out := suite.out.result

	suite.NotEmpty(out)
	suite.Contains(out, "help")
	suite.Nil(err)
}

func (suite *ExecutorSuite) TestRunWithArgZsh() {
	err := suite.exec.run(suite.cmd, []string{"zsh"})
	out := suite.out.result

	suite.NotEmpty(out)
	suite.Contains(out, "BASH_COMPLETION_EOF")
	suite.Contains(out, "convert_bash_to_zsh")
	suite.Nil(err)
}

func (suite *ExecutorSuite) TestRunWithArgBash() {
	err := suite.exec.run(suite.cmd, []string{"bash"})
	out := suite.out.result

	suite.NotEmpty(out)
	suite.Contains(out, "_init_completion")
	suite.Nil(err)
}

func (suite *ExecutorSuite) TestRunWithBadArg() {
	err := suite.exec.run(suite.cmd, []string{"fish"})
	out := suite.out.result

	suite.NotEmpty(out)
	suite.Contains(out, "unknown shell")
	suite.Contains(out, "usage")
	suite.Nil(err)
}

func (suite *ExecutorSuite) TestHelpWithNoArgs() {
	suite.exec.runHelp(suite.cmd, []string{"completion"})
	out := suite.out.result

	suite.NotEmpty(out)
	suite.Contains(out, "Output shell completion code for the given shell")
	suite.Contains(out, "help")
}

func (suite *ExecutorSuite) TestHelpWithZsh() {
	suite.exec.runHelp(suite.cmd, []string{"completion", "zsh"})
	out := suite.out.result

	suite.NotEmpty(out)
	suite.Contains(out, "Add the following")
	suite.Contains(out, "zshrc")
}

func (suite *ExecutorSuite) TestHelpWithBash() {
	suite.exec.runHelp(suite.cmd, []string{"completion", "bash"})
	out := suite.out.result

	suite.NotEmpty(out)
	suite.Contains(out, "add the following")
	suite.Contains(out, "bash_profile")
}

func (suite *ExecutorSuite) TestHelpWithBadArg() {
	suite.exec.runHelp(suite.cmd, []string{"completion", "fish"})
	out := suite.out.result

	suite.NotEmpty(out)
	suite.Contains(out, "unknown shell")
	suite.Contains(out, "help")
}

func TestRunExecSuite(t *testing.T) {
	suite.Run(t, new(ExecutorSuite))
}

type exWriter struct {
	result string
}

func (w *exWriter) Clean() {
	w.result = ""
}

func (w *exWriter) Write(p []byte) (int, error) {
	w.result += string(p)
	return 0, nil
}

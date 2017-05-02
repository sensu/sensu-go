package commands

import (
	"testing"

	"github.com/sensu/sensu-go/cli/commands/test"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestAddCommands(t *testing.T) {
	assert := assert.New(t)
	cmd := &cobra.Command{}
	cli := test.SimpleSensuCLI(nil)

	AddCommands(cmd, cli)
	assert.NotEmpty(cmd.Commands())
}

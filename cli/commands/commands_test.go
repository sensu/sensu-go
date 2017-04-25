package commands

import (
	"testing"

	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestAddCommands(t *testing.T) {
	assert := assert.New(t)
	cmd := &cobra.Command{}
	cli := &cli.SensuCli{}

	AddCommands(cmd, cli)
	assert.NotEmpty(cmd.Commands())
}

package configure

import (
	"testing"

	"github.com/sensu/sensu-go/cli"
	"github.com/stretchr/testify/assert"
)

func TestCommand(t *testing.T) {
	assert := assert.New(t)

	cli := &cli.SensuCli{}
	cmd := Command(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.Run, "cmd should be able to be executed")
	assert.Regexp("configure", cmd.Use)
	assert.Regexp("Initialize sensuctl configuration", cmd.Short)
}

package cluster

import (
	"testing"

	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
)

func TestHealthCommand(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewCLI()
	cmd := HealthCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("health", cmd.Use)
	assert.Regexp("get sensu health status", cmd.Short)
}

package agent

import (
	"path/filepath"
	"testing"

	"github.com/sensu/sensu-go/testing/testutil"
	"github.com/sensu/sensu-go/transport"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestExecuteHook(t *testing.T) {
	assert := assert.New(t)

	hookConfig := types.FixtureHookConfig("hook")
	hookConfig.Stdin = true

	config := FixtureConfig()
	agent := NewAgent(config)
	ch := make(chan *transport.Message, 1)
	agent.sendq = ch

	truePath := testutil.CommandPath(filepath.Join(toolsDir, "true"))
	hookConfig.Command = truePath

	hook := agent.executeHook(hookConfig)

	assert.NotZero(hook.Executed)
	assert.Equal(hook.Status, int32(0))
	assert.Equal(hook.Output, "")

	hookConfig.Command = "printf hello"

	hook = agent.executeHook(hookConfig)

	assert.NotZero(hook.Executed)
	assert.Equal(hook.Status, int32(0))
	assert.Equal(hook.Output, "hello")
}

func TestPrepareHook(t *testing.T) {
	assert := assert.New(t)

	config := FixtureConfig()
	agent := NewAgent(config)

	// nil hook
	assert.False(agent.prepareHook(nil))

	// Invalid hook
	hook := types.FixtureHookConfig("hook")
	hook.Command = ""
	assert.False(agent.prepareHook(hook))

	// Valid check
	hook.Command = "{{ .ID }}"
	assert.True(agent.prepareHook(hook))
}

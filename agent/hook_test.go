package agent

import (
	"context"
	"errors"
	"testing"

	time "github.com/echlebek/timeproxy"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/command"
	"github.com/sensu/sensu-go/testing/mockexecutor"

	"github.com/sensu/sensu-go/transport"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestExecuteHook(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	hookConfig := types.FixtureHookConfig("hook")
	hookConfig.Stdin = true

	config, cleanup := FixtureConfig()
	defer cleanup()
	agent, err := NewAgent(config)
	if err != nil {
		t.Fatal(err)
	}
	ch := make(chan *transport.Message, 1)
	agent.sendq = ch
	ex := &mockexecutor.MockExecutor{}
	agent.executor = ex
	execution := command.FixtureExecutionResponse(0, "")
	ex.Return(execution, nil)

	evt := &types.Event{
		Check: &types.Check{
			ObjectMeta: types.ObjectMeta{
				Name: "check",
			},
		},
	}

	hook := agent.executeHook(ctx, hookConfig, evt, nil)

	assert.NotZero(hook.Executed)
	assert.Equal(int32(0), hook.Status)
	assert.Equal("", hook.Output)

	execution.Output = "hello"
	hook = agent.executeHook(ctx, hookConfig, evt, nil)

	assert.NotZero(hook.Executed)
	assert.Equal(int32(0), hook.Status)
	assert.Equal("hello", hook.Output)
}

func TestExecuteHooks_GH3779(t *testing.T) {
	cfg, cleanup := FixtureConfig()
	defer cleanup()
	ag, err := NewAgent(cfg)
	if err != nil {
		t.Fatal(err)
	}
	request := &corev2.CheckRequest{
		Config: corev2.FixtureCheckConfig("foo"),
		// Deliberately set hooks to nil
		Hooks:  nil,
		Issued: time.Now().Unix(),
	}
	request.Config.CheckHooks = []corev2.HookList{
		{
			Hooks: []string{
				"doesnotexist",
			},
			Type: "ok",
		},
	}
	event := corev2.FixtureEvent("foo", "foo")
	assets := make(map[string]*corev2.AssetList)
	hooks := ag.ExecuteHooks(context.Background(), request, event, assets)
	if got, want := len(hooks), 1; got != want {
		t.Fatal("expected 1 hook")
	}
}

func TestPrepareHook(t *testing.T) {
	config, cleanup := FixtureConfig()
	defer cleanup()
	agent, err := NewAgent(config)
	if err != nil {
		t.Fatal(err)
	}

	// nil hook
	if err := agent.prepareHook(nil); err == nil {
		t.Error("expected non-nil error")
	}

	// Invalid hook
	hook := types.FixtureHookConfig("hook")
	hook.Command = ""
	if err := agent.prepareHook(hook); err == nil {
		t.Error("expected non-nil error")
	}

	// Valid check
	hook.Command = "{{ .name }}"
	if err := agent.prepareHook(hook); err != nil {
		t.Error(err)
	}
}

func TestHookInList(t *testing.T) {
	assert := assert.New(t)
	hook1 := types.FixtureHook("hook1")
	hook2 := types.FixtureHook("hook2")

	testCases := []struct {
		name     string
		hookName string
		hookList []*types.Hook
		expected bool
	}{
		{
			name:     "Empty list",
			hookName: "hook1",
			hookList: []*types.Hook{},
			expected: false,
		},
		{
			name:     "Hook in populated list",
			hookName: "hook1",
			hookList: []*types.Hook{hook2, hook1},
			expected: true,
		},
		{
			name:     "Hook not in populated list",
			hookName: "hook1",
			hookList: []*types.Hook{hook2, hook2},
			expected: false,
		},
		{
			name:     "No hook name provided",
			hookName: "",
			hookList: []*types.Hook{hook1, hook2},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			in := hookInList(tc.hookName, tc.hookList)
			assert.Equal(tc.expected, in)
		})
	}
}

func TestErrorHookConfig(t *testing.T) {
	hc := errorHookConfig("default", "agent", errors.New("it ain't work"))
	if err := hc.Validate(); err != nil {
		t.Fatal(err)
	}
}

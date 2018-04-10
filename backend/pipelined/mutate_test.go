// Package pipelined provides the traditional Sensu event pipeline.
package pipelined

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/sensu/sensu-go/rpc"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHelperMutatorProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_MUTATOR_PROCESS") != "1" {
		return
	}

	command := strings.Join(os.Args[3:], " ")
	stdin, _ := ioutil.ReadAll(os.Stdin)

	switch command {
	case "cat":
		fmt.Fprintf(os.Stdout, "%s", stdin)
	}
	os.Exit(0)
}

func TestPipelinedMutate(t *testing.T) {
	p, err := New(Config{Store: nil, Bus: nil})
	require.NoError(t, err)

	handler := types.FakeHandlerCommand("cat")
	handler.Type = "pipe"

	event := &types.Event{}

	eventData, err := p.mutateEvent(handler, event)

	expected, _ := json.Marshal(event)

	assert.NoError(t, err)
	assert.Equal(t, expected, eventData)
}

func TestPipelinedJsonMutator(t *testing.T) {
	p, err := New(Config{Store: nil, Bus: nil})
	require.NoError(t, err)

	event := &types.Event{}

	output, err := p.jsonMutator(event)

	expected, _ := json.Marshal(event)

	assert.NoError(t, err)
	assert.Equal(t, expected, output)
}

func TestPipelinedOnlyCheckOutputMutator(t *testing.T) {
	p, err := New(Config{Store: nil, Bus: nil})
	require.NoError(t, err)

	event := &types.Event{}
	event.Check = &types.Check{}
	event.Check.Output = "foo"

	output := p.onlyCheckOutputMutator(event)

	expected := []byte("foo")
	assert.Equal(t, expected, output)
}

func TestPipelinedOnlyCheckOutputMutate(t *testing.T) {
	p, err := New(Config{Store: nil, Bus: nil})
	require.NoError(t, err)

	handler := types.FakeHandlerCommand("cat")
	handler.Type = "pipe"
	handler.Mutator = "only_check_output"

	event := &types.Event{}
	event.Check = &types.Check{}
	event.Check.Output = "foo"

	eventData, err := p.mutateEvent(handler, event)

	expected := []byte("foo")

	assert.NoError(t, err)
	assert.Equal(t, expected, eventData)
}

func TestPipelinedExtensionMutator(t *testing.T) {
	m := &mockExec{}
	ext := &types.Extension{URL: "http://127.0.0.1"}
	store := &mockstore.MockStore{}
	store.On("GetExtension", mock.Anything, "extension").Return(ext, nil)
	store.On("GetMutatorByName", mock.Anything, "extension").Return((*types.Mutator)(nil), nil)
	event := types.FixtureEvent("foo", "bar")

	m.On("MutateEvent", event).Return([]byte("remote"), nil)

	getter := func(*types.Extension) (rpc.ExtensionExecutor, error) {
		return m, nil
	}

	p, err := New(Config{ExtensionExecutorGetter: getter, Store: store})
	require.NoError(t, err)

	handler := &types.Handler{}
	handler.Mutator = "extension"

	eventData, err := p.mutateEvent(handler, event)
	require.NoError(t, err)
	require.Equal(t, []byte("remote"), eventData)
}

func TestPipelinedPipeMutator(t *testing.T) {
	p, err := New(Config{Store: nil, Bus: nil})
	require.NoError(t, err)

	mutator := types.FakeMutatorCommand("cat")

	event := &types.Event{}

	output, err := p.pipeMutator(mutator, event)

	expected, _ := json.Marshal(event)

	assert.NoError(t, err)
	assert.Equal(t, expected, output)
}

package pipeline

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/secrets"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/rpc"
	"github.com/sensu/sensu-go/testing/mockstore"
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

func TestPipelineMutate(t *testing.T) {
	p := New(Config{})

	handler := corev2.FakeHandlerCommand("cat")
	handler.Type = "pipe"

	event := &corev2.Event{}

	eventData, err := p.mutateEvent(handler, event)

	expected, _ := json.Marshal(event)

	assert.NoError(t, err)
	assert.Equal(t, expected, eventData)
}

func TestPipelineJsonMutator(t *testing.T) {
	p := New(Config{})

	event := &corev2.Event{}

	output, err := p.jsonMutator(event)

	expected, _ := json.Marshal(event)

	assert.NoError(t, err)
	assert.Equal(t, expected, output)
}

func TestPipelineOnlyCheckOutputMutator(t *testing.T) {
	p := New(Config{})

	event := &corev2.Event{}
	event.Check = &corev2.Check{}
	event.Check.Output = "foo"

	output := p.onlyCheckOutputMutator(event)

	expected := []byte("foo")
	assert.Equal(t, expected, output)
}

func TestPipelineOnlyCheckOutputMutate(t *testing.T) {
	p := New(Config{})

	handler := corev2.FakeHandlerCommand("cat")
	handler.Type = "pipe"
	handler.Mutator = "only_check_output"

	event := &corev2.Event{}
	event.Check = &corev2.Check{}
	event.Check.Output = "foo"

	eventData, err := p.mutateEvent(handler, event)

	expected := []byte("foo")

	assert.NoError(t, err)
	assert.Equal(t, expected, eventData)
}

func TestPipelineExtensionMutator(t *testing.T) {
	m := &mockExec{}
	ext := &corev2.Extension{URL: "http://127.0.0.1"}
	store := &mockstore.MockStore{}
	store.On("GetExtension", mock.Anything, "extension").Return(ext, nil)
	store.On("GetMutatorByName", mock.Anything, "extension").Return((*corev2.Mutator)(nil), nil)
	event := corev2.FixtureEvent("foo", "bar")

	m.On("MutateEvent", event).Return([]byte("remote"), nil)

	getter := func(*corev2.Extension) (rpc.ExtensionExecutor, error) {
		return m, nil
	}

	p := New(Config{ExtensionExecutorGetter: getter, Store: store})

	handler := &corev2.Handler{}
	handler.Mutator = "extension"

	eventData, err := p.mutateEvent(handler, event)
	require.NoError(t, err)
	require.Equal(t, []byte("remote"), eventData)
}

func TestPipelinePipeMutator(t *testing.T) {
	p := New(Config{SecretsProviderManager: secrets.NewProviderManager()})

	mutator := corev2.FakeMutatorCommand("cat")

	event := &corev2.Event{}

	output, err := p.pipeMutator(mutator, event)

	expected, _ := json.Marshal(event)

	assert.NoError(t, err)
	assert.Equal(t, expected, output)
}

func TestPipelineNoMutator_GH2784(t *testing.T) {
	stor := &mockstore.MockStore{}
	stor.On("GetExtension", mock.Anything, mock.Anything).Return((*corev2.Extension)(nil), store.ErrNoExtension)
	stor.On("GetMutatorByName", mock.Anything, mock.Anything).Return((*corev2.Mutator)(nil), nil)

	event := corev2.FixtureEvent("foo", "bar")
	handler := &corev2.Handler{Mutator: "nope"}

	p := New(Config{Store: stor})
	eventData, err := p.mutateEvent(handler, event)
	require.Error(t, err)
	require.Nil(t, eventData)
}

package pipeline

import (
	"encoding/json"
	"io"
	"io/ioutil"
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

	output, err := p.pipeMutator(mutator, event, nil)

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

func TestPipelineJavascriptMutatorImplicit(t *testing.T) {
	mutator := &corev2.Mutator{
		ObjectMeta: corev2.ObjectMeta{
			Namespace: "default",
			Name:      "my_mutator",
		},
		Eval:    `event.check.labels["hockey"] = hockey;`,
		Type:    corev2.JavascriptMutator,
		EnvVars: []string{"hockey=puck"},
	}

	handler := &corev2.Handler{
		ObjectMeta: corev2.ObjectMeta{
			Namespace: "default",
			Name:      "my_handler",
		},
		Mutator: "my_mutator",
	}

	stor := &mockstore.MockStore{}
	stor.On("GetExtension", mock.Anything, mock.Anything).Return((*corev2.Extension)(nil), store.ErrNoExtension)
	stor.On("GetMutatorByName", mock.Anything, mock.Anything).Return(mutator, nil)

	p := New(Config{SecretsProviderManager: secrets.NewProviderManager(), Store: stor})

	event := corev2.FixtureEvent("default", "default")

	output, err := p.mutateEvent(handler, event)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := event.Check.Labels["hockey"], "puck"; got != want {
		t.Errorf("bad mutation: got %q, want %q", got, want)
	}

	expected, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	assert.JSONEq(t, string(output), string(expected))
}

func TestPipelineJavascriptMutatorExplicit(t *testing.T) {
	mutator := &corev2.Mutator{
		ObjectMeta: corev2.ObjectMeta{
			Namespace: "default",
			Name:      "my_mutator",
		},
		Eval:    `event.check.labels["hockey"] = hockey; return JSON.stringify(event);`,
		Type:    corev2.JavascriptMutator,
		EnvVars: []string{"hockey=puck"},
	}

	handler := &corev2.Handler{
		ObjectMeta: corev2.ObjectMeta{
			Namespace: "default",
			Name:      "my_handler",
		},
		Mutator: "my_mutator",
	}

	stor := &mockstore.MockStore{}
	stor.On("GetExtension", mock.Anything, mock.Anything).Return((*corev2.Extension)(nil), store.ErrNoExtension)
	stor.On("GetMutatorByName", mock.Anything, mock.Anything).Return(mutator, nil)

	p := New(Config{SecretsProviderManager: secrets.NewProviderManager(), Store: stor})

	event := corev2.FixtureEvent("default", "default")

	output, err := p.mutateEvent(handler, event)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := event.Check.Labels["hockey"], "puck"; got != want {
		t.Errorf("bad mutation: got %q, want %q", got, want)
	}

	expected, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	assert.JSONEq(t, string(output), string(expected))
}

func TestPipelineJavascriptMutatorObjectFailure(t *testing.T) {
	mutator := &corev2.Mutator{
		ObjectMeta: corev2.ObjectMeta{
			Namespace: "default",
			Name:      "my_mutator",
		},
		Eval:    `event.check.labels["hockey"] = hockey; return {};`,
		Type:    corev2.JavascriptMutator,
		EnvVars: []string{"hockey=puck"},
	}

	handler := &corev2.Handler{
		ObjectMeta: corev2.ObjectMeta{
			Namespace: "default",
			Name:      "my_handler",
		},
		Mutator: "my_mutator",
	}

	stor := &mockstore.MockStore{}
	stor.On("GetExtension", mock.Anything, mock.Anything).Return((*corev2.Extension)(nil), store.ErrNoExtension)
	stor.On("GetMutatorByName", mock.Anything, mock.Anything).Return(mutator, nil)

	p := New(Config{SecretsProviderManager: secrets.NewProviderManager(), Store: stor})

	event := corev2.FixtureEvent("default", "default")

	if _, err := p.mutateEvent(handler, event); err == nil {
		t.Fatal("expected non-nil error")
	}
}

func TestPipelineJavascriptMutatorTimeoutFailure(t *testing.T) {
	mutator := &corev2.Mutator{
		ObjectMeta: corev2.ObjectMeta{
			Namespace: "default",
			Name:      "my_mutator",
		},
		Eval:    `sleep(2);`,
		Type:    corev2.JavascriptMutator,
		EnvVars: []string{"hockey=puck"},
		Timeout: 1,
	}

	handler := &corev2.Handler{
		ObjectMeta: corev2.ObjectMeta{
			Namespace: "default",
			Name:      "my_handler",
		},
		Mutator: "my_mutator",
	}

	stor := &mockstore.MockStore{}
	stor.On("GetExtension", mock.Anything, mock.Anything).Return((*corev2.Extension)(nil), store.ErrNoExtension)
	stor.On("GetMutatorByName", mock.Anything, mock.Anything).Return(mutator, nil)

	p := New(Config{SecretsProviderManager: secrets.NewProviderManager(), Store: stor})

	event := corev2.FixtureEvent("default", "default")

	if _, err := p.mutateEvent(handler, event); err == nil {
		t.Fatal("expected non-nil error")
	}
}

func TestPipelineJavascriptMutatorReturnNull(t *testing.T) {
	mutator := &corev2.Mutator{
		ObjectMeta: corev2.ObjectMeta{
			Namespace: "default",
			Name:      "my_mutator",
		},
		Eval:    `event.check.labels["hockey"] = hockey; return null;`,
		Type:    corev2.JavascriptMutator,
		EnvVars: []string{"hockey=puck"},
	}

	handler := &corev2.Handler{
		ObjectMeta: corev2.ObjectMeta{
			Namespace: "default",
			Name:      "my_handler",
		},
		Mutator: "my_mutator",
	}

	stor := &mockstore.MockStore{}
	stor.On("GetExtension", mock.Anything, mock.Anything).Return((*corev2.Extension)(nil), store.ErrNoExtension)
	stor.On("GetMutatorByName", mock.Anything, mock.Anything).Return(mutator, nil)

	p := New(Config{SecretsProviderManager: secrets.NewProviderManager(), Store: stor})

	event := corev2.FixtureEvent("default", "default")

	output, err := p.mutateEvent(handler, event)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := event.Check.Labels["hockey"], "puck"; got != want {
		t.Errorf("bad mutation: got %q, want %q", got, want)
	}

	expected, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	assert.JSONEq(t, string(output), string(expected))
}

type mutatorAssetSet struct {
}

func (mutatorAssetSet) Key() string {
	return "mutatorAsset"
}

func (mutatorAssetSet) Scripts() (map[string]io.ReadCloser, error) {
	result := make(map[string]io.ReadCloser)
	result["mutatorAsset"] = ioutil.NopCloser(strings.NewReader(`var assetFunc = function () { event.check.labels["hockey"] = hockey; }`))
	return result, nil
}

func TestPipelineJavascriptMutatorUseAsset(t *testing.T) {
	mutator := &corev2.Mutator{
		ObjectMeta: corev2.ObjectMeta{
			Namespace: "default",
			Name:      "my_mutator",
		},
		Eval:          `assetFunc();`,
		Type:          corev2.JavascriptMutator,
		EnvVars:       []string{"hockey=puck"},
		RuntimeAssets: []string{"mutatorAsset"},
	}

	stor := &mockstore.MockStore{}
	stor.On("GetExtension", mock.Anything, mock.Anything).Return((*corev2.Extension)(nil), store.ErrNoExtension)
	stor.On("GetMutatorByName", mock.Anything, mock.Anything).Return(mutator, nil)

	p := New(Config{SecretsProviderManager: secrets.NewProviderManager(), Store: stor})

	event := corev2.FixtureEvent("default", "default")

	output, err := p.javascriptMutator(mutator, event, mutatorAssetSet{})
	if err != nil {
		t.Fatal(err)
	}

	if got, want := event.Check.Labels["hockey"], "puck"; got != want {
		t.Errorf("bad mutation: got %q, want %q", got, want)
	}

	expected, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	assert.JSONEq(t, string(output), string(expected))
}

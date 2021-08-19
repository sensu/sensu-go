package pipeline

import (
	"encoding/json"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/secrets"
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

func TestPipelinePipeMutator(t *testing.T) {
	p := New(Config{SecretsProviderManager: secrets.NewProviderManager()})

	mutator := corev2.FakeMutatorCommand("cat")

	event := &corev2.Event{}

	output, err := p.pipeMutator(mutator, event, nil)

	expected, _ := json.Marshal(event)

	assert.NoError(t, err)
	assert.Equal(t, expected, output)
}

// test that missing mutators returns an error
func TestPipelineNoMutator_GH2784(t *testing.T) {
	stor := &mockstore.MockStore{}
	stor.On("GetMutatorByName", mock.Anything, mock.Anything).Return((*corev2.Mutator)(nil), nil)

	event := corev2.FixtureEvent("foo", "bar")
	handler := &corev2.Handler{Mutator: "nope"}

	p := New(Config{Store: stor})
	eventData, err := p.mutateEvent(handler, event)
	require.Error(t, err)
	require.Nil(t, eventData)
}

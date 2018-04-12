package graphql

import (
	"context"
	"errors"
	"testing"

	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type eventFinder struct {
	collection []*types.Event
	err        error
}

func (f eventFinder) Query(ctx context.Context, entity, check string) ([]*types.Event, error) {
	return f.collection, f.err
}

func TestEnvColourID(t *testing.T) {
	handler := &envImpl{}
	env := types.Environment{Name: "pink"}

	colour, err := handler.ColourID(graphql.ResolveParams{Source: &env})
	assert.NoError(t, err)
	assert.Equal(t, string(colour), "BLUE")
}

func TestEnvironmentTypeCheckHistoryField(t *testing.T) {
	env := types.Environment{Name: "pink"}
	events := []*types.Event{
		types.FixtureEvent("a", "b"),
		types.FixtureEvent("b", "c"),
		types.FixtureEvent("c", "d"),
	}

	finder := eventFinder{collection: events, err: nil}
	impl := &envImpl{eventsCtrl: finder}

	// Params
	params := schema.EnvironmentCheckHistoryFieldResolverParams{}
	params.Source = &env

	// limit: 30
	params.Args.Limit = 30
	history, err := impl.CheckHistory(params)
	require.NoError(t, err)
	assert.NotEmpty(t, history)
	assert.Len(t, history, 30)

	// store err
	impl.eventsCtrl = eventFinder{err: errors.New("test")}
	history, err = impl.CheckHistory(params)
	require.NotNil(t, history)
	assert.Error(t, err)
	assert.Empty(t, history)
}

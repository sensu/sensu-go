package graphql

import (
	"errors"
	"testing"

	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvColourID(t *testing.T) {
	impl := &envImpl{}
	env := types.Environment{Name: "pink"}

	colour, err := impl.ColourID(graphql.ResolveParams{Source: &env})
	assert.NoError(t, err)
	assert.Equal(t, string(colour), "BLUE")
}

func TestEnvironmentTypeCheckHistoryField(t *testing.T) {
	mock := mockEventQuerier{els: []*types.Event{
		types.FixtureEvent("a", "b"),
		types.FixtureEvent("b", "c"),
		types.FixtureEvent("c", "d"),
	}}
	impl := &envImpl{eventQuerier: mock}

	// Params
	params := schema.EnvironmentCheckHistoryFieldResolverParams{}
	params.Source = &types.Environment{Name: "pink"}

	// limit: 30
	params.Args.Limit = 30
	history, err := impl.CheckHistory(params)
	require.NoError(t, err)
	assert.NotEmpty(t, history)
	assert.Len(t, history, 30)

	// store err
	impl.eventQuerier = mockEventQuerier{err: errors.New("test")}
	history, err = impl.CheckHistory(params)
	require.NotNil(t, history)
	assert.Error(t, err)
	assert.Empty(t, history)
}

func TestEnvironmentTypeSilencesField(t *testing.T) {
	mock := mockSilenceQuerier{els: []*types.Silenced{
		types.FixtureSilenced("a:b"),
		types.FixtureSilenced("b:c"),
		types.FixtureSilenced("c:d"),
	}}
	impl := &envImpl{silenceQuerier: mock}

	// Params
	params := schema.EnvironmentSilencesFieldResolverParams{}
	params.Source = types.FixtureEnvironment("xxx")

	// Success
	res, err := impl.Silences(params)
	require.NoError(t, err)
	assert.NotEmpty(t, res)

	// Store err
	impl.silenceQuerier = mockSilenceQuerier{err: errors.New("test")}
	res, err = impl.Silences(params)
	assert.Empty(t, res)
	assert.Error(t, err)
}

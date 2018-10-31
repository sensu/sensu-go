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

func TestNamespaceTypeColourID(t *testing.T) {
	impl := &namespaceImpl{}
	nsp := types.Namespace{Name: "pink"}

	colour, err := impl.ColourID(graphql.ResolveParams{Source: &nsp})
	assert.NoError(t, err)
	assert.Equal(t, string(colour), "BLUE")
}

func TestNamespaceTypeCheckHistoryField(t *testing.T) {
	mock := mockEventQuerier{els: []*types.Event{
		types.FixtureEvent("a", "b"),
		types.FixtureEvent("b", "c"),
		types.FixtureEvent("c", "d"),
	}}
	impl := &namespaceImpl{eventQuerier: mock}

	// Params
	params := schema.NamespaceCheckHistoryFieldResolverParams{}
	params.Source = &types.Namespace{Name: "pink"}

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

func TestNamespaceTypeSilencesField(t *testing.T) {
	mock := mockSilenceQuerier{els: []*types.Silenced{
		types.FixtureSilenced("a:b"),
		types.FixtureSilenced("b:c"),
		types.FixtureSilenced("c:d"),
	}}
	impl := &namespaceImpl{silenceQuerier: mock}

	// Params
	params := schema.NamespaceSilencesFieldResolverParams{}
	params.Source = types.FixtureNamespace("xxx")

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

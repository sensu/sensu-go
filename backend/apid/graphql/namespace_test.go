package graphql

import (
	"errors"
	"testing"

	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	client "github.com/sensu/sensu-go/backend/apid/graphql/testing"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
	client, factory := client.NewClientFactory()
	client.On("ListEvents", mock.Anything).Return([]types.Event{
		*types.FixtureEvent("a", "b"),
		*types.FixtureEvent("b", "c"),
		*types.FixtureEvent("c", "d"),
	}, nil).Once()
	impl := &namespaceImpl{factory}

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
	client.On("ListEvents", mock.Anything).Return([]types.Event{}, errors.New("test"))
	history, err = impl.CheckHistory(params)
	require.NotNil(t, history)
	assert.Error(t, err)
	assert.Empty(t, history)
}

func TestNamespaceTypeSilencesField(t *testing.T) {
	client, factory := client.NewClientFactory()
	client.On("ListSilenceds", mock.Anything, "", "").Return([]types.Silenced{
		*types.FixtureSilenced("a:b"),
		*types.FixtureSilenced("b:c"),
		*types.FixtureSilenced("c:d"),
	}, nil).Once()
	impl := &namespaceImpl{factory}

	// Params
	params := schema.NamespaceSilencesFieldResolverParams{}
	params.Source = types.FixtureNamespace("xxx")

	// Success
	res, err := impl.Silences(params)
	require.NoError(t, err)
	assert.NotEmpty(t, res)

	// Store err
	client.On("ListSilenceds", mock.Anything, "", "").Return([]types.Silenced{}, errors.New("abc"))
	res, err = impl.Silences(params)
	assert.Empty(t, res)
	assert.Error(t, err)
}

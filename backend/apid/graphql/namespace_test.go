package graphql

import (
	"context"
	"errors"
	"testing"

	client "github.com/sensu/sensu-go/backend/apid/graphql/mockclient"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNamespaceTypeColourID(t *testing.T) {
	impl := &namespaceImpl{}
	nsp := types.Namespace{Name: "sensu"}

	colour, err := impl.ColourID(graphql.ResolveParams{Source: &nsp})
	assert.NoError(t, err)
	assert.Equal(t, string(colour), "ORANGE")
}

func TestNamespaceTypeSilencesField(t *testing.T) {
	client, _ := client.NewClientFactory()
	client.On("ListSilenceds", mock.Anything, "", "", mock.Anything).Return([]types.Silenced{
		*types.FixtureSilenced("a:b"),
		*types.FixtureSilenced("b:c"),
		*types.FixtureSilenced("c:d"),
	}, nil).Once()

	impl := &namespaceImpl{}
	params := schema.NamespaceSilencesFieldResolverParams{}
	params.Context = contextWithLoadersNoCache(context.Background(), client)
	params.Source = types.FixtureNamespace("xxx")

	// Success
	res, err := impl.Silences(params)
	require.NoError(t, err)
	assert.NotEmpty(t, res)

	// Store err
	client.On("ListSilenceds", mock.Anything, "", "", mock.Anything).Return([]types.Silenced{}, errors.New("abc"))
	res, err = impl.Silences(params)
	assert.Empty(t, res)
	assert.Error(t, err)
}

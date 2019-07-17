package graphql

import (
	"context"
	"testing"

	v2 "github.com/sensu/sensu-go/api/core/v2"
	client "github.com/sensu/sensu-go/backend/apid/graphql/mockclient"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandlerTypeHandlersField(t *testing.T) {
	handler := types.FixtureHandler("my-handler")
	handler.Handlers = []string{"one", "two"}

	client, factory := client.NewClientFactory()
	impl := &handlerImpl{}

	params := graphql.ResolveParams{}
	cfg := ServiceConfig{ClientFactory: factory}
	params.Context = contextWithLoadersNoCache(context.Background(), cfg)
	params.Source = handler

	// Success
	client.On("ListHandlers", mock.Anything, mock.Anything).Return([]types.Handler{
		*types.FixtureHandler("one"),
		*types.FixtureHandler("two"),
		*types.FixtureHandler("three"),
	}, nil).Once()

	res, err := impl.Handlers(params)
	require.NoError(t, err)
	assert.NotEmpty(t, res)
}

func TestHandlerTypeMutatorField(t *testing.T) {
	mutator := types.FixtureMutator("my-mutator")
	handler := types.FixtureHandler("my-handler")
	handler.Mutator = mutator.Name

	client, factory := client.NewClientFactory()
	impl := &handlerImpl{factory: factory}

	// Success
	client.On("FetchMutator", mutator.Name).Return(mutator, nil).Once()
	res, err := impl.Mutator(graphql.ResolveParams{Source: handler})
	require.NoError(t, err)
	assert.NotEmpty(t, res)

	// No mutator
	handler.Mutator = ""
	res, err = impl.Mutator(graphql.ResolveParams{Source: handler})
	require.NoError(t, err)
	assert.Nil(t, res)
}

func TestHandlerTypeToJSONField(t *testing.T) {
	src := v2.FixtureHandler("name")
	imp := &handlerImpl{}

	res, err := imp.ToJSON(graphql.ResolveParams{Source: src})
	require.NoError(t, err)
	assert.NotEmpty(t, res)
}

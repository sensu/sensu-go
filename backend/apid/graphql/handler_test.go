package graphql

import (
	"context"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandlerTypeHandlersField(t *testing.T) {
	handler := corev2.FixtureHandler("my-handler")
	handler.Handlers = []string{"one", "two", "four", "six:seven"}

	client := new(MockGenericClient)
	client.On("List", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		list := args.Get(1).(*[]*corev2.Handler)
		*list = []*corev2.Handler{
			corev2.FixtureHandler("one"),
			corev2.FixtureHandler("two"),
			corev2.FixtureHandler("three"),
			corev2.FixtureHandler("four:five"),
			corev2.FixtureHandler("six:seven"),
		}
	}).Return(nil).Once()

	cfg := ServiceConfig{GenericClient: client}

	params := graphql.ResolveParams{}
	params.Context = contextWithLoadersNoCache(context.Background(), cfg)
	params.Source = handler

	resolver := &handlerImpl{}
	got, err := resolver.Handlers(params)
	require.NoError(t, err)
	assert.Len(t, got, 3)
}

func TestHandlerTypeMutatorField(t *testing.T) {
	mutator := corev2.FixtureMutator("my-mutator")
	handler := corev2.FixtureHandler("my-handler")
	handler.Mutator = mutator.Name

	client := new(MockMutatorClient)
	impl := &handlerImpl{client: client}

	// Success
	client.On("FetchMutator", mock.Anything, mutator.Name).Return(mutator, nil).Once()
	res, err := impl.Mutator(graphql.ResolveParams{Source: handler, Context: context.Background()})
	require.NoError(t, err)
	assert.NotEmpty(t, res)

	// No mutator
	handler.Mutator = ""
	res, err = impl.Mutator(graphql.ResolveParams{Source: handler, Context: context.Background()})
	require.NoError(t, err)
	assert.Nil(t, res)
}

func TestHandlerTypeToJSONField(t *testing.T) {
	src := corev2.FixtureHandler("name")
	imp := &handlerImpl{}

	res, err := imp.ToJSON(graphql.ResolveParams{Source: src, Context: context.Background()})
	require.NoError(t, err)
	assert.NotEmpty(t, res)
}

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

func TestSilencedTypeCheckField(t *testing.T) {
	check := corev2.FixtureCheckConfig("http-check")
	silenced := corev2.FixtureSilenced("unix:http-check")

	client := new(MockCheckClient)
	impl := &silencedImpl{client: client}

	// Success
	client.On("FetchCheck", mock.Anything, check.Name).Return(check, nil).Once()
	res, err := impl.Check(graphql.ResolveParams{Source: silenced, Context: context.Background()})
	require.NoError(t, err)
	assert.NotEmpty(t, res)
}

func TestSilencedTypeExpiresField(t *testing.T) {
	silenced := corev2.FixtureSilenced("unix:http-check")
	impl := &silencedImpl{}

	// with expiry unset
	res, err := impl.Expires(graphql.ResolveParams{Source: silenced, Context: context.Background()})
	require.NoError(t, err)
	assert.Nil(t, res)

	// with expiry set
	silenced.ExpireAt = 1234
	res, err = impl.Expires(graphql.ResolveParams{Source: silenced, Context: context.Background()})
	require.NoError(t, err)
	assert.NotNil(t, res)
}

func TestSilencedTypeExpireAtField(t *testing.T) {
	silenced := corev2.FixtureSilenced("unix:http-check")
	impl := &silencedImpl{}

	// with expiry unset
	res, err := impl.ExpireAt(graphql.ResolveParams{Source: silenced, Context: context.Background()})
	require.NoError(t, err)
	assert.Nil(t, res)

	// with expiry set
	silenced.ExpireAt = 1234
	res, err = impl.ExpireAt(graphql.ResolveParams{Source: silenced, Context: context.Background()})
	require.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, convertTs(1234), res)
}

func TestSilencedTypeBeginField(t *testing.T) {
	silenced := corev2.FixtureSilenced("unix:http-check")
	impl := &silencedImpl{}

	res, err := impl.Begin(graphql.ResolveParams{Source: silenced, Context: context.Background()})
	require.NoError(t, err)
	assert.Nil(t, res)
}

func TestSilencedTypeToJSONField(t *testing.T) {
	src := corev2.FixtureSilenced("check:subscription")
	imp := &silencedImpl{}

	res, err := imp.ToJSON(graphql.ResolveParams{Source: src, Context: context.Background()})
	require.NoError(t, err)
	assert.NotEmpty(t, res)
}

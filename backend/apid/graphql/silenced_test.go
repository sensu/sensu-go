package graphql

import (
	"testing"

	client "github.com/sensu/sensu-go/backend/apid/graphql/mockclient"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSilencedTypeCheckField(t *testing.T) {
	check := types.FixtureCheckConfig("http-check")
	silenced := types.FixtureSilenced("unix:http-check")

	client, factory := client.NewClientFactory()
	impl := &silencedImpl{factory: factory}

	// Success
	client.On("FetchCheck", check.Name).Return(check, nil).Once()
	res, err := impl.Check(graphql.ResolveParams{Source: silenced})
	require.NoError(t, err)
	assert.NotEmpty(t, res)
}

func TestSilencedTypeExpiresField(t *testing.T) {
	silenced := types.FixtureSilenced("unix:http-check")
	impl := &silencedImpl{}

	// with expiry unset
	res, err := impl.Expires(graphql.ResolveParams{Source: silenced})
	require.NoError(t, err)
	assert.Nil(t, res)

	// with expiry set
	silenced.Expire = 1234
	res, err = impl.Expires(graphql.ResolveParams{Source: silenced})
	require.NoError(t, err)
	assert.NotNil(t, res)
}

func TestSilencedTypeBeginField(t *testing.T) {
	silenced := types.FixtureSilenced("unix:http-check")
	impl := &silencedImpl{}

	res, err := impl.Begin(graphql.ResolveParams{Source: silenced})
	require.NoError(t, err)
	assert.Nil(t, res)
}

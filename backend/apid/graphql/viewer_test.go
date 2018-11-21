package graphql

import (
	"context"
	"testing"

	client "github.com/sensu/sensu-go/backend/apid/graphql/testing"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
	clientlib "github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestViewerTypeEventField(t *testing.T) {
	client, factory := client.NewClientFactory()
	impl := viewerImpl{factory: factory}

	user := types.FixtureUser("frankwest")
	claims, err := jwt.NewClaims(user)
	if err != nil {
		require.NoError(t, err)
	}

	params := graphql.ResolveParams{}
	params.Context = context.WithValue(context.Background(), types.ClaimsKey, claims)

	// Success
	client.On("FetchUser", user.Username).Return(user, nil).Once()
	res, err := impl.User(params)
	require.NoError(t, err)
	assert.NotEmpty(t, res)

	// User not found for claim
	client.On("FetchUser", user.Username).Return(user, clientlib.APIError{Code: 2}).Once()
	res, err = impl.User(params)
	require.NoError(t, err)
	assert.Empty(t, res)

	// No claims
	params.Context = context.Background()
	res, err = impl.User(params)
	require.NoError(t, err)
	assert.Empty(t, res)
}

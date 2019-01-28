package graphql

import (
	"context"
	"testing"

	mockclient "github.com/sensu/sensu-go/backend/apid/graphql/mockclient"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestViewerTypeUserField(t *testing.T) {
	client, factory := mockclient.NewClientFactory()
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
	client.On("FetchUser", user.Username).Return(user, mockclient.NotFound).Once()
	res, err = impl.User(params)
	require.NoError(t, err)
	assert.Empty(t, res)

	// No claims
	params.Context = context.Background()
	res, err = impl.User(params)
	require.NoError(t, err)
	assert.Empty(t, res)
}

func TestViewerTypeNamespacesField(t *testing.T) {
	nsp := types.FixtureNamespace("sensu")
	impl := viewerImpl{}
	client, _ := mockclient.NewClientFactory()

	params := graphql.ResolveParams{}
	params.Context = contextWithLoadersNoCache(context.Background(), client)

	// Success
	client.On("ListNamespaces").Return([]types.Namespace{*nsp}, nil).Once()
	res, err := impl.Namespaces(params)
	require.NoError(t, err)
	assert.NotEmpty(t, res)
}

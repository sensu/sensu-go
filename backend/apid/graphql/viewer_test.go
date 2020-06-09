package graphql

import (
	"context"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestViewerTypeUserField(t *testing.T) {
	client := new(MockUserClient)
	impl := viewerImpl{userClient: client}

	user := corev2.FixtureUser("frankwest")
	claims, err := jwt.NewClaims(user)
	if err != nil {
		require.NoError(t, err)
	}

	params := graphql.ResolveParams{}
	params.Context = context.WithValue(context.Background(), corev2.ClaimsKey, claims)

	// Success
	client.On("FetchUser", mock.Anything, user.Username).Return(user, nil).Once()
	res, err := impl.User(params)
	require.NoError(t, err)
	assert.NotEmpty(t, res)

	// User not found for claim
	client.On("FetchUser", mock.Anything, user.Username).Return(user, &store.ErrNotFound{}).Once()
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
	nsp := corev2.FixtureNamespace("sensu")
	impl := viewerImpl{}
	client := new(MockNamespaceClient)

	params := graphql.ResolveParams{}
	cfg := ServiceConfig{NamespaceClient: client}
	params.Context = contextWithLoadersNoCache(context.Background(), cfg)

	// Success
	client.On("ListNamespaces", mock.Anything, mock.Anything).Return([]*corev2.Namespace{nsp}, nil).Once()
	res, err := impl.Namespaces(params)
	require.NoError(t, err)
	assert.NotEmpty(t, res)
}

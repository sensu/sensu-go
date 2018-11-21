package graphql

import (
	"context"
	"fmt"
	"testing"

	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	mockclient "github.com/sensu/sensu-go/backend/apid/graphql/mockclient"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func setupNodeResolver(factory ClientFactory) func(string) (interface{}, error) {
	resolver := newNodeResolver(factory)
	ctx := context.Background()
	info := graphql.ResolveInfo{}

	return func(gid string) (interface{}, error) {
		return resolver.Find(ctx, gid, info)
	}
}

func TestNodeResolverFindType(t *testing.T) {
	_, factory := mockclient.NewClientFactory()
	resolver := newNodeResolver(factory)

	check := types.FixtureCheckConfig("http-check")
	typeID := resolver.FindType(check)
	assert.NotNil(t, typeID)
}

func TestNodeResolverFind(t *testing.T) {
	client, factory := mockclient.NewClientFactory()
	resolver := newNodeResolver(factory)

	ctx := context.Background()
	info := graphql.ResolveInfo{}

	check := types.FixtureCheckConfig("http-check")
	gid := globalid.CheckTranslator.EncodeToString(check)

	// Sucess
	client.On("FetchCheck", check.Name).Return(check, nil).Once()
	res, err := resolver.Find(ctx, gid, info)
	assert.NotEmpty(t, res)
	assert.NoError(t, err)

	// Missing
	client.On("FetchCheck", check.Name).Return(check, mockclient.NotFound).Once()
	res, err = resolver.Find(ctx, gid, info)
	assert.Empty(t, res)
	assert.NoError(t, err)

	// Error
	client.On("FetchCheck", check.Name).Return(check, mockclient.InternalErr).Once()
	res, err = resolver.Find(ctx, gid, info)
	assert.Empty(t, res)
	assert.Error(t, err)

	// Bad ID
	res, err = resolver.Find(ctx, "sadfasdfasdf", info)
	assert.Empty(t, res)
	assert.Error(t, err)
}

func TestAssetNodeResolver(t *testing.T) {
	client, factory := mockclient.NewClientFactory()
	find := setupNodeResolver(factory)

	testCases := []struct {
		name      string
		setupNode func() interface{}
		setupID   func(interface{}) string
		setup     func(interface{})
	}{
		{
			name: "assets",
			setupNode: func() interface{} {
				return types.FixtureAsset("name")
			},
			setupID: func(r interface{}) string {
				return globalid.AssetTranslator.EncodeToString(r)
			},
			setup: func(r interface{}) {
				client.On("FetchAsset", "name").Return(r, nil).Once()
			},
		},
		{
			name: "check",
			setupNode: func() interface{} {
				return types.FixtureCheckConfig("name")
			},
			setupID: func(r interface{}) string {
				return globalid.CheckTranslator.EncodeToString(r)
			},
			setup: func(r interface{}) {
				client.On("FetchCheck", "name").Return(r, nil).Once()
			},
		},
		{
			name: "entities",
			setupNode: func() interface{} {
				return types.FixtureEntity("name")
			},
			setupID: func(r interface{}) string {
				return globalid.EntityTranslator.EncodeToString(r)
			},
			setup: func(r interface{}) {
				client.On("FetchEntity", "name").Return(r, nil).Once()
			},
		},
		{
			name: "handlers",
			setupNode: func() interface{} {
				return types.FixtureHandler("name")
			},
			setupID: func(r interface{}) string {
				return globalid.HandlerTranslator.EncodeToString(r)
			},
			setup: func(r interface{}) {
				client.On("FetchHandler", "name").Return(r, nil).Once()
			},
		},
		{
			name: "mutators",
			setupNode: func() interface{} {
				return types.FixtureMutator("name")
			},
			setupID: func(r interface{}) string {
				return globalid.MutatorTranslator.EncodeToString(r)
			},
			setup: func(r interface{}) {
				client.On("FetchMutator", "name").Return(r, nil).Once()
			},
		},
		{
			name: "users",
			setupNode: func() interface{} {
				return types.FixtureUser("name")
			},
			setupID: func(r interface{}) string {
				return globalid.UserTranslator.EncodeToString(r)
			},
			setup: func(r interface{}) {
				client.On("FetchUser", "name").Return(r, nil).Once()
			},
		},
		{
			name: "events",
			setupNode: func() interface{} {
				return types.FixtureEvent("a", "b")
			},
			setupID: func(r interface{}) string {
				return globalid.EventTranslator.EncodeToString(r)
			},
			setup: func(r interface{}) {
				client.On("FetchEvent", "a", "b").Return(r, nil).Once()
			},
		},
		{
			name: "namespaces",
			setupNode: func() interface{} {
				return types.FixtureNamespace("sensu")
			},
			setupID: func(r interface{}) string {
				return globalid.NamespaceTranslator.EncodeToString(r)
			},
			setup: func(r interface{}) {
				client.On("FetchNamespace", "sensu").Return(r, nil).Once()
			},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("can resolve %s", tc.name), func(t *testing.T) {
			in := tc.setupNode()
			id := tc.setupID(in)
			tc.setup(in)

			res, err := find(id)
			assert.Equal(t, res, in)
			assert.NoError(t, err)
		})
	}
}

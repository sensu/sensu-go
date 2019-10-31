package graphql

import (
	"context"
	"errors"
	"fmt"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupNodeResolver(cfg ServiceConfig) func(string) (interface{}, error) {
	resolver := newNodeResolver(cfg)
	ctx := context.Background()
	info := graphql.ResolveInfo{}

	return func(gid string) (interface{}, error) {
		return resolver.Find(ctx, gid, info)
	}
}

func TestNodeResolverFindType(t *testing.T) {
	cfg := ServiceConfig{}
	resolver := newNodeResolver(cfg)

	check := corev2.FixtureCheckConfig("http-check")
	typeID := resolver.FindType(context.Background(), check)
	assert.NotNil(t, typeID)
}

func TestNodeResolverFind(t *testing.T) {
	client := new(MockCheckClient)
	cfg := ServiceConfig{CheckClient: client}
	resolver := newNodeResolver(cfg)

	ctx := context.Background()
	info := graphql.ResolveInfo{}

	check := corev2.FixtureCheckConfig("http-check")
	gid := globalid.CheckTranslator.EncodeToString(context.Background(), check)

	// Success
	client.On("FetchCheck", mock.Anything, check.Name).Return(check, nil).Once()
	res, err := resolver.Find(ctx, gid, info)
	assert.NotEmpty(t, res)
	assert.NoError(t, err)

	// Missing
	client.On("FetchCheck", mock.Anything, check.Name).Return(check, &store.ErrNotFound{}).Once()
	res, err = resolver.Find(ctx, gid, info)
	assert.Empty(t, res)
	assert.NoError(t, err)

	// Error
	client.On("FetchCheck", mock.Anything, check.Name).Return(check, errors.New("an error")).Once()
	res, err = resolver.Find(ctx, gid, info)
	assert.Empty(t, res)
	assert.Error(t, err)

	// Bad ID
	res, err = resolver.Find(ctx, "sadfasdfasdf", info)
	assert.Empty(t, res)
	assert.Error(t, err)
}

type onner interface {
	On(string, ...interface{}) *mock.Call
}

func TestAssetNodeResolver(t *testing.T) {
	cfg := ServiceConfig{
		AssetClient:       new(MockAssetClient),
		CheckClient:       new(MockCheckClient),
		EntityClient:      new(MockEntityClient),
		EventClient:       new(MockEventClient),
		EventFilterClient: new(MockEventFilterClient),
		HandlerClient:     new(MockHandlerClient),
		MutatorClient:     new(MockMutatorClient),
		UserClient:        new(MockUserClient),
		NamespaceClient:   new(MockNamespaceClient),
	}
	find := setupNodeResolver(cfg)

	testCases := []struct {
		name      string
		setupNode func() interface{}
		setupID   func(interface{}) string
		setup     func(interface{})
	}{
		{
			name: "assets",
			setupNode: func() interface{} {
				return corev2.FixtureAsset("name")
			},
			setupID: func(r interface{}) string {
				return globalid.AssetTranslator.EncodeToString(context.Background(), r)
			},
			setup: func(r interface{}) {
				cfg.AssetClient.(onner).On("FetchAsset", mock.Anything, "name").Return(r, nil).Once()
			},
		},
		{
			name: "check",
			setupNode: func() interface{} {
				return corev2.FixtureCheckConfig("name")
			},
			setupID: func(r interface{}) string {
				return globalid.CheckTranslator.EncodeToString(context.Background(), r)
			},
			setup: func(r interface{}) {
				cfg.CheckClient.(onner).On("FetchCheck", mock.Anything, "name").Return(r, nil).Once()
			},
		},
		{
			name: "entities",
			setupNode: func() interface{} {
				return corev2.FixtureEntity("name")
			},
			setupID: func(r interface{}) string {
				return globalid.EntityTranslator.EncodeToString(context.Background(), r)
			},
			setup: func(r interface{}) {
				cfg.EntityClient.(onner).On("FetchEntity", mock.Anything, "name").Return(r, nil).Once()
			},
		},
		{
			name: "event filters",
			setupNode: func() interface{} {
				return corev2.FixtureEventFilter("name")
			},
			setupID: func(r interface{}) string {
				return globalid.EventFilterTranslator.EncodeToString(context.Background(), r)
			},
			setup: func(r interface{}) {
				cfg.EventFilterClient.(onner).On("FetchEventFilter", mock.Anything, "name").Return(r, nil).Once()
			},
		},
		{
			name: "handlers",
			setupNode: func() interface{} {
				return corev2.FixtureHandler("name")
			},
			setupID: func(r interface{}) string {
				return globalid.HandlerTranslator.EncodeToString(context.Background(), r)
			},
			setup: func(r interface{}) {
				cfg.HandlerClient.(onner).On("FetchHandler", mock.Anything, "name").Return(r, nil).Once()
			},
		},
		{
			name: "mutators",
			setupNode: func() interface{} {
				return corev2.FixtureMutator("name")
			},
			setupID: func(r interface{}) string {
				return globalid.MutatorTranslator.EncodeToString(context.Background(), r)
			},
			setup: func(r interface{}) {
				cfg.MutatorClient.(onner).On("FetchMutator", mock.Anything, "name").Return(r, nil).Once()
			},
		},
		{
			name: "users",
			setupNode: func() interface{} {
				return corev2.FixtureUser("name")
			},
			setupID: func(r interface{}) string {
				return globalid.UserTranslator.EncodeToString(context.Background(), r)
			},
			setup: func(r interface{}) {
				cfg.UserClient.(onner).On("FetchUser", mock.Anything, "name").Return(r, nil).Once()
			},
		},
		{
			name: "events",
			setupNode: func() interface{} {
				return corev2.FixtureEvent("a", "b")
			},
			setupID: func(r interface{}) string {
				return globalid.EventTranslator.EncodeToString(context.Background(), r)
			},
			setup: func(r interface{}) {
				cfg.EventClient.(onner).On("FetchEvent", mock.Anything, "a", "b").Return(r, nil).Once()
			},
		},
		{
			name: "namespaces",
			setupNode: func() interface{} {
				return corev2.FixtureNamespace("sensu")
			},
			setupID: func(r interface{}) string {
				return globalid.NamespaceTranslator.EncodeToString(context.Background(), r)
			},
			setup: func(r interface{}) {
				cfg.NamespaceClient.(onner).On("FetchNamespace", mock.Anything, "sensu").Return(r, nil).Once()
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

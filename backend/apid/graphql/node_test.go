package graphql

import (
	"context"
	"fmt"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/apid/graphql/relay"
	"github.com/sensu/sensu-go/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupNodeResolver(cfg ServiceConfig) func(string) (interface{}, error) {
	register := relay.NodeRegister{}
	resolver := relay.Resolver{Register: &register}
	registerNodeResolvers(register, cfg)

	ctx := context.Background()
	info := graphql.ResolveInfo{}

	return func(gid string) (interface{}, error) {
		return resolver.Find(ctx, gid, info)
	}
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

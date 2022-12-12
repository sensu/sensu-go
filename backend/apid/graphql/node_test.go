package graphql

import (
	"context"
	"fmt"
	"testing"

	corev2 "github.com/sensu/core/v2"
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

func TestNodeResolvers(t *testing.T) {
	cfg := ServiceConfig{
		EntityClient:      new(MockEntityClient),
		EventClient:       new(MockEventClient),
		EventFilterClient: new(MockEventFilterClient),
		GenericClient:     new(MockGenericClient),
		HandlerClient:     new(MockHandlerClient),
		MutatorClient:     new(MockMutatorClient),
		NamespaceClient:   new(MockNamespaceClient),
		RBACClient:        new(MockRBACClient),
		SilencedClient:    new(MockSilencedClient),
		UserClient:        new(MockUserClient),
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
				cfg.GenericClient.(onner).On("SetTypeMeta", mock.Anything).Return(nil).Once()
				cfg.GenericClient.(onner).On("Get", mock.Anything, "name", mock.Anything).Run(func(args mock.Arguments) {
					arg := args.Get(2).(*corev2.Asset)
					*arg = *r.(*corev2.Asset)
				}).Return(nil).Once()
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
				cfg.GenericClient.(onner).On("SetTypeMeta", mock.Anything).Return(nil).Once()
				cfg.GenericClient.(onner).On("Get", mock.Anything, "name", mock.Anything).Run(func(args mock.Arguments) {
					arg := args.Get(2).(*corev2.CheckConfig)
					*arg = *r.(*corev2.CheckConfig)
				}).Return(nil).Once()
			},
		},
		{
			name: "entities",
			setupNode: func() interface{} {
				return corev2.FixtureEntity("sensu")
			},
			setupID: func(r interface{}) string {
				return globalid.EntityTranslator.EncodeToString(context.Background(), r)
			},
			setup: func(r interface{}) {
				cfg.EntityClient.(onner).On("FetchEntity", mock.Anything, "sensu").Return(r, nil).Once()
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
				cfg.GenericClient.(onner).On("SetTypeMeta", mock.Anything).Return(nil).Once()
				cfg.GenericClient.(onner).On("Get", mock.Anything, "name", mock.Anything).Run(func(args mock.Arguments) {
					arg := args.Get(2).(*corev2.EventFilter)
					*arg = *r.(*corev2.EventFilter)
				}).Return(nil).Once()
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
				cfg.GenericClient.(onner).On("SetTypeMeta", mock.Anything).Return(nil).Once()
				cfg.GenericClient.(onner).On("Get", mock.Anything, "name", mock.Anything).Run(func(args mock.Arguments) {
					arg := args.Get(2).(*corev2.Handler)
					*arg = *r.(*corev2.Handler)
				}).Return(nil).Once()
			},
		},
		{
			name: "hooks",
			setupNode: func() interface{} {
				return corev2.FixtureHookConfig("name")
			},
			setupID: func(r interface{}) string {
				return globalid.HookTranslator.EncodeToString(context.Background(), r)
			},
			setup: func(r interface{}) {
				cfg.GenericClient.(onner).On("SetTypeMeta", mock.Anything).Return(nil).Once()
				cfg.GenericClient.(onner).On("Get", mock.Anything, "name", mock.Anything).Run(func(args mock.Arguments) {
					arg := args.Get(2).(*corev2.HookConfig)
					*arg = *r.(*corev2.HookConfig)
				}).Return(nil).Once()
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
				cfg.GenericClient.(onner).On("SetTypeMeta", mock.Anything).Return(nil).Once()
				cfg.GenericClient.(onner).On("Get", mock.Anything, "name", mock.Anything).Run(func(args mock.Arguments) {
					arg := args.Get(2).(*corev2.Mutator)
					*arg = *r.(*corev2.Mutator)
				}).Return(nil).Once()
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
		{
			name: "cluster-role",
			setupNode: func() interface{} {
				return corev2.FixtureClusterRole("sensu")
			},
			setupID: func(r interface{}) string {
				return globalid.ClusterRoleTranslator.EncodeToString(context.Background(), r)
			},
			setup: func(r interface{}) {
				cfg.RBACClient.(onner).On("FetchClusterRole", mock.Anything, "sensu").Return(r, nil).Once()
			},
		},
		{
			name: "cluster-role-binding",
			setupNode: func() interface{} {
				return corev2.FixtureClusterRoleBinding("sensu")
			},
			setupID: func(r interface{}) string {
				return globalid.ClusterRoleBindingTranslator.EncodeToString(context.Background(), r)
			},
			setup: func(r interface{}) {
				cfg.RBACClient.(onner).On("FetchClusterRoleBinding", mock.Anything, "sensu").Return(r, nil).Once()
			},
		},
		{
			name: "role",
			setupNode: func() interface{} {
				return corev2.FixtureRole("sensu", "default")
			},
			setupID: func(r interface{}) string {
				return globalid.RoleTranslator.EncodeToString(context.Background(), r)
			},
			setup: func(r interface{}) {
				cfg.RBACClient.(onner).On("FetchRole", mock.Anything, "sensu").Return(r, nil).Once()
			},
		},
		{
			name: "role-binding",
			setupNode: func() interface{} {
				return corev2.FixtureRoleBinding("sensu", "default")
			},
			setupID: func(r interface{}) string {
				return globalid.RoleBindingTranslator.EncodeToString(context.Background(), r)
			},
			setup: func(r interface{}) {
				cfg.RBACClient.(onner).On("FetchRoleBinding", mock.Anything, "sensu").Return(r, nil).Once()
			},
		},
		{
			name: "silenced",
			setupNode: func() interface{} {
				return corev2.FixtureSilenced("sub:check")
			},
			setupID: func(r interface{}) string {
				return globalid.SilenceTranslator.EncodeToString(context.Background(), r)
			},
			setup: func(r interface{}) {
				cfg.SilencedClient.(onner).On("GetSilencedByName", mock.Anything, "sub:check").Return(r, nil).Once()
			},
		},
		{
			name: "pipeline",
			setupNode: func() interface{} {
				return corev2.FixturePipeline("name", "default")
			},
			setupID: func(r interface{}) string {
				return globalid.PipelineTranslator.EncodeToString(context.Background(), r)
			},
			setup: func(r interface{}) {
				cfg.GenericClient.(onner).On("SetTypeMeta", mock.Anything).Return(nil).Once()
				cfg.GenericClient.(onner).On("Get", mock.Anything, "name", mock.Anything).Run(func(args mock.Arguments) {
					arg := args.Get(2).(*corev2.Pipeline)
					*arg = *r.(*corev2.Pipeline)
				}).Return(nil).Once()
			},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("can resolve %s", tc.name), func(t *testing.T) {
			in := tc.setupNode()
			id := tc.setupID(in)
			tc.setup(in)

			res, err := find(id)
			assert.Equal(t, in, res)
			assert.NoError(t, err)
		})
	}
}

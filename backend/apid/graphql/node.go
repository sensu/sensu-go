package graphql

import (
	"errors"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/apid/graphql/relay"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	util_relay "github.com/sensu/sensu-go/backend/apid/graphql/util/relay"
)

func registerNodeResolvers(register relay.NodeRegister, cfg ServiceConfig) {
	registerAssetNodeResolver(register, cfg.GenericClient)
	registerCheckNodeResolver(register, cfg.GenericClient)
	registerClusterRoleNodeResolver(register, cfg.RBACClient)
	registerClusterRoleBindingNodeResolver(register, cfg.RBACClient)
	registerEntityNodeResolver(register, cfg.EntityClient)
	registerEventNodeResolver(register, cfg.EventClient)
	registerEventFilterNodeResolver(register, cfg.GenericClient)
	registerHandlerNodeResolver(register, cfg.GenericClient)
	registerHookNodeResolver(register, cfg.GenericClient)
	registerMutatorNodeResolver(register, cfg.GenericClient)
	registerNamespaceNodeResolver(register, cfg.NamespaceClient)
	registerPipelineNodeResolver(register, cfg.GenericClient)
	registerRoleNodeResolver(register, cfg.RBACClient)
	registerRoleBindingNodeResolver(register, cfg.RBACClient)
	registerUserNodeResolver(register, cfg.UserClient)
	registerSilencedNodeResolver(register, cfg.SilencedClient)
	register.RegisterResolver(relay.NodeResolver{
		Translator: GlobalIDCoreV3EntityConfig,
		ObjectType: schema.CoreV3EntityConfigType,
		Resolve: util_relay.MakeNodeResolver(
			cfg.GenericClient,
			corev2.TypeMeta{Type: "EntityConfig", APIVersion: "core/v3"},
		),
	})
	register.RegisterResolver(relay.NodeResolver{
		Translator: GlobalIDCoreV3EntityState,
		ObjectType: schema.CoreV3EntityStateType,
		Resolve: util_relay.MakeNodeResolver(
			cfg.GenericClient,
			corev2.TypeMeta{Type: "EntityState", APIVersion: "core/v3"},
		),
	})
}

// assets

func registerAssetNodeResolver(register relay.NodeRegister, client GenericClient) {
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.AssetType,
		Translator: globalid.AssetTranslator,
		Resolve: util_relay.MakeNodeResolver(
			client,
			corev2.TypeMeta{Type: "Asset", APIVersion: "core/v2"}),
	})
}

// checks

func registerCheckNodeResolver(register relay.NodeRegister, client GenericClient) {
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.CheckConfigType,
		Translator: globalid.CheckTranslator,
		Resolve: util_relay.MakeNodeResolver(
			client,
			corev2.TypeMeta{Type: "CheckConfig", APIVersion: "core/v2"}),
	})
}

// entities

func registerEntityNodeResolver(register relay.NodeRegister, client EntityClient) {
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.EntityType,
		Translator: globalid.EntityTranslator,
		Resolve: func(p relay.NodeResolverParams) (interface{}, error) {
			ctx := setContextFromComponents(p.Context, p.IDComponents)
			record, err := client.FetchEntity(ctx, p.IDComponents.UniqueComponent())
			return handleFetchResult(record, err)
		},
	})
}

// event filters

func registerEventFilterNodeResolver(register relay.NodeRegister, client GenericClient) {
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.EventFilterType,
		Translator: globalid.EventFilterTranslator,
		Resolve: util_relay.MakeNodeResolver(
			client,
			corev2.TypeMeta{Type: "EventFilter", APIVersion: "core/v2"}),
	})
}

// handlers

func registerHandlerNodeResolver(register relay.NodeRegister, client GenericClient) {
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.HandlerType,
		Translator: globalid.HandlerTranslator,
		Resolve: util_relay.MakeNodeResolver(
			client,
			corev2.TypeMeta{Type: "Handler", APIVersion: "core/v2"}),
	})
}

// hooks

func registerHookNodeResolver(register relay.NodeRegister, client GenericClient) {
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.HookConfigType,
		Translator: globalid.HookTranslator,
		Resolve: util_relay.MakeNodeResolver(
			client,
			corev2.TypeMeta{Type: "HookConfig", APIVersion: "core/v2"}),
	})
}

// mutators

func registerMutatorNodeResolver(register relay.NodeRegister, client GenericClient) {
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.MutatorType,
		Translator: globalid.MutatorTranslator,
		Resolve: util_relay.MakeNodeResolver(
			client,
			corev2.TypeMeta{Type: "Mutator", APIVersion: "core/v2"}),
	})
}

// cluster roles

func registerClusterRoleNodeResolver(register relay.NodeRegister, client RBACClient) {
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.ClusterRoleType,
		Translator: globalid.ClusterRoleTranslator,
		Resolve: func(p relay.NodeResolverParams) (interface{}, error) {
			ctx := setContextFromComponents(p.Context, p.IDComponents)
			record, err := client.FetchClusterRole(ctx, p.IDComponents.UniqueComponent())
			return handleFetchResult(record, err)
		},
	})
}

// cluster role bindings

func registerClusterRoleBindingNodeResolver(register relay.NodeRegister, client RBACClient) {
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.ClusterRoleBindingType,
		Translator: globalid.ClusterRoleBindingTranslator,
		Resolve: func(p relay.NodeResolverParams) (interface{}, error) {
			ctx := setContextFromComponents(p.Context, p.IDComponents)
			record, err := client.FetchClusterRoleBinding(ctx, p.IDComponents.UniqueComponent())
			return handleFetchResult(record, err)
		},
	})
}

// roles

func registerRoleNodeResolver(register relay.NodeRegister, client RBACClient) {
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.RoleType,
		Translator: globalid.RoleTranslator,
		Resolve: func(p relay.NodeResolverParams) (interface{}, error) {
			ctx := setContextFromComponents(p.Context, p.IDComponents)
			record, err := client.FetchRole(ctx, p.IDComponents.UniqueComponent())
			return handleFetchResult(record, err)
		},
	})
}

// role bindings

func registerRoleBindingNodeResolver(register relay.NodeRegister, client RBACClient) {
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.RoleBindingType,
		Translator: globalid.RoleBindingTranslator,
		Resolve: func(p relay.NodeResolverParams) (interface{}, error) {
			ctx := setContextFromComponents(p.Context, p.IDComponents)
			record, err := client.FetchRoleBinding(ctx, p.IDComponents.UniqueComponent())
			return handleFetchResult(record, err)
		},
	})
}

// user

type userNodeResolver struct {
	client UserClient
}

func registerUserNodeResolver(register relay.NodeRegister, client UserClient) {
	resolver := &userNodeResolver{client: client}
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.UserType,
		Translator: globalid.UserTranslator,
		Resolve:    resolver.fetch,
	})
}

func (f *userNodeResolver) fetch(p relay.NodeResolverParams) (interface{}, error) {
	ctx := setContextFromComponents(p.Context, p.IDComponents)
	record, err := f.client.FetchUser(ctx, p.IDComponents.UniqueComponent())
	return handleFetchResult(record, err)
}

// events

type eventNodeResolver struct {
	client EventClient
}

func registerEventNodeResolver(register relay.NodeRegister, client EventClient) {
	resolver := &eventNodeResolver{client: client}
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.EventType,
		Translator: globalid.EventTranslator,
		Resolve:    resolver.fetch,
	})
}

func (f *eventNodeResolver) fetch(p relay.NodeResolverParams) (interface{}, error) {
	evComponents, ok := p.IDComponents.(globalid.EventComponents)
	if !ok {
		return nil, errors.New("given id does not appear to reference event")
	}

	ctx := setContextFromComponents(p.Context, p.IDComponents)
	record, err := f.client.FetchEvent(ctx, evComponents.EntityName(), evComponents.CheckName())
	return handleFetchResult(record, err)
}

// namespaces

type namespaceNodeResolver struct {
	client NamespaceClient
}

func registerNamespaceNodeResolver(register relay.NodeRegister, client NamespaceClient) {
	resolver := &namespaceNodeResolver{client: client}
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.NamespaceType,
		Translator: globalid.NamespaceTranslator,
		Resolve:    resolver.fetch,
	})
}

func (f *namespaceNodeResolver) fetch(p relay.NodeResolverParams) (interface{}, error) {
	record, err := f.client.FetchNamespace(p.Context, p.IDComponents.UniqueComponent())
	return handleFetchResult(record, err)
}

// silences

func registerSilencedNodeResolver(register relay.NodeRegister, client SilencedClient) {
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.SilencedType,
		Translator: globalid.SilenceTranslator,
		Resolve: func(p relay.NodeResolverParams) (interface{}, error) {
			ctx := setContextFromComponents(p.Context, p.IDComponents)
			record, err := client.GetSilencedByName(ctx, p.IDComponents.UniqueComponent())
			return handleFetchResult(record, err)
		},
	})
}

// pipelines

func registerPipelineNodeResolver(register relay.NodeRegister, client GenericClient) {
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.CoreV2PipelineType,
		Translator: globalid.PipelineTranslator,
		Resolve: util_relay.MakeNodeResolver(
			client,
			corev2.TypeMeta{Type: "Pipeline", APIVersion: "core/v2"}),
	})
}

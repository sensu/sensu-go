package graphql

import (
	"context"
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/apid/graphql/relay"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/graphql"
)

//
// Node Resolver
//

type nodeResolver struct {
	register *relay.NodeRegister
}

func newNodeResolver(cfg ServiceConfig) *nodeResolver {
	register := relay.NodeRegister{}

	registerAssetNodeResolver(register, cfg.AssetClient)
	registerCheckNodeResolver(register, cfg.CheckClient)
	registerEntityNodeResolver(register, cfg.EntityClient)
	registerEventFilterNodeResolver(register, cfg.EventFilterClient)
	registerHandlerNodeResolver(register, cfg.HandlerClient)
	registerHookNodeResolver(register, cfg.HookClient)
	registerMutatorNodeResolver(register, cfg.MutatorClient)
	registerClusterRoleNodeResolver(register, cfg.RBACClient)
	registerClusterRoleBindingNodeResolver(register, cfg.RBACClient)
	registerRoleNodeResolver(register, cfg.RBACClient)
	registerRoleBindingNodeResolver(register, cfg.RBACClient)
	registerUserNodeResolver(register, cfg.UserClient)
	registerEventNodeResolver(register, cfg.EventClient)
	registerNamespaceNodeResolver(register, cfg.NamespaceClient)
	registerSilencedNodeResolver(register, cfg.SilencedClient)

	return &nodeResolver{&register}
}

func (r *nodeResolver) FindType(ctx context.Context, i interface{}) *graphql.Type {
	translator, err := globalid.ReverseLookup(i)
	if err != nil {
		return nil
	}

	components := translator.Encode(ctx, i)
	resolver := r.register.Lookup(components)
	if resolver == nil {
		logger := logger.WithField("translator", fmt.Sprintf("%#v", translator))
		logger.Error("unable to find node resolver for type")
		return nil
	}
	return &resolver.ObjectType
}

func (r *nodeResolver) Find(ctx context.Context, id string, info graphql.ResolveInfo) (interface{}, error) {
	// Decode given ID
	idComponents, err := globalid.Decode(id)
	if err != nil {
		return nil, err
	}

	// Lookup resolver using components of a global ID
	resolver := r.register.Lookup(idComponents)
	if resolver == nil {
		return nil, errors.New("unable to find type associated with this ID")
	}

	// Lift org & env into context
	ctx = setContextFromComponents(ctx, idComponents)

	// Fetch resource from using resolver
	params := relay.NodeResolverParams{
		Context:      ctx,
		IDComponents: idComponents,
		Info:         info,
	}
	return resolver.Resolve(params)
}

// assets

type assetNodeResolver struct {
	client AssetClient
}

func registerAssetNodeResolver(register relay.NodeRegister, client AssetClient) {
	resolver := &assetNodeResolver{client: client}
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.AssetType,
		Translator: globalid.AssetTranslator,
		Resolve:    resolver.fetch,
	})
}

func (f *assetNodeResolver) fetch(p relay.NodeResolverParams) (interface{}, error) {
	ctx := setContextFromComponents(p.Context, p.IDComponents)
	record, err := f.client.FetchAsset(ctx, p.IDComponents.UniqueComponent())
	return handleFetchResult(record, err)
}

// checks

type checkNodeResolver struct {
	client CheckClient
}

func registerCheckNodeResolver(register relay.NodeRegister, client CheckClient) {
	resolver := &checkNodeResolver{client: client}
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.CheckConfigType,
		Translator: globalid.CheckTranslator,
		Resolve:    resolver.fetch,
	})
}

func (f *checkNodeResolver) fetch(p relay.NodeResolverParams) (interface{}, error) {
	ctx := setContextFromComponents(p.Context, p.IDComponents)
	record, err := f.client.FetchCheck(ctx, p.IDComponents.UniqueComponent())
	return handleFetchResult(record, err)
}

// entities

type entityNodeResolver struct {
	client EntityClient
}

func registerEntityNodeResolver(register relay.NodeRegister, client EntityClient) {
	resolver := &entityNodeResolver{client: client}
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.EntityType,
		Translator: globalid.EntityTranslator,
		Resolve:    resolver.fetch,
	})
}

func (f *entityNodeResolver) fetch(p relay.NodeResolverParams) (interface{}, error) {
	ctx := setContextFromComponents(p.Context, p.IDComponents)
	record, err := f.client.FetchEntity(ctx, p.IDComponents.UniqueComponent())
	return handleFetchResult(record, err)
}

// event filters

type eventFilterNodeResolver struct {
	client EventFilterClient
}

func registerEventFilterNodeResolver(register relay.NodeRegister, client EventFilterClient) {
	resolver := &eventFilterNodeResolver{client: client}
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.EventFilterType,
		Translator: globalid.EventFilterTranslator,
		Resolve:    resolver.fetch,
	})
}

func (f *eventFilterNodeResolver) fetch(p relay.NodeResolverParams) (interface{}, error) {
	ctx := setContextFromComponents(p.Context, p.IDComponents)
	record, err := f.client.FetchEventFilter(ctx, p.IDComponents.UniqueComponent())
	return handleFetchResult(record, err)
}

// handlers

type handlerNodeResolver struct {
	client HandlerClient
}

func registerHandlerNodeResolver(register relay.NodeRegister, client HandlerClient) {
	resolver := &handlerNodeResolver{client: client}
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.HandlerType,
		Translator: globalid.HandlerTranslator,
		Resolve:    resolver.fetch,
	})
}

func (f *handlerNodeResolver) fetch(p relay.NodeResolverParams) (interface{}, error) {
	ctx := setContextFromComponents(p.Context, p.IDComponents)
	record, err := f.client.FetchHandler(ctx, p.IDComponents.UniqueComponent())
	return handleFetchResult(record, err)
}

// hooks

type hookNodeResolver struct {
	client HookClient
}

func registerHookNodeResolver(register relay.NodeRegister, client HookClient) {
	resolver := &hookNodeResolver{client: client}
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.HookConfigType,
		Translator: globalid.HookTranslator,
		Resolve:    resolver.fetch,
	})
}

func (f *hookNodeResolver) fetch(p relay.NodeResolverParams) (interface{}, error) {
	ctx := setContextFromComponents(p.Context, p.IDComponents)
	record, err := f.client.FetchHookConfig(ctx, p.IDComponents.UniqueComponent())
	return handleFetchResult(record, err)
}

// mutators

type mutatorNodeResolver struct {
	client MutatorClient
}

func registerMutatorNodeResolver(register relay.NodeRegister, client MutatorClient) {
	resolver := &mutatorNodeResolver{client: client}
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.MutatorType,
		Translator: globalid.MutatorTranslator,
		Resolve:    resolver.fetch,
	})
}

func (f *mutatorNodeResolver) fetch(p relay.NodeResolverParams) (interface{}, error) {
	ctx := setContextFromComponents(p.Context, p.IDComponents)
	record, err := f.client.FetchMutator(ctx, p.IDComponents.UniqueComponent())
	return handleFetchResult(record, err)
}

// cluster roles

type clusterRoleNodeResolver struct {
	client RBACClient
}

func registerClusterRoleNodeResolver(register relay.NodeRegister, client RBACClient) {
	resolver := &clusterRoleNodeResolver{client: client}
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.ClusterRoleType,
		Translator: globalid.ClusterRoleTranslator,
		Resolve:    resolver.fetch,
	})
}

func (f *clusterRoleNodeResolver) fetch(p relay.NodeResolverParams) (interface{}, error) {
	ctx := setContextFromComponents(p.Context, p.IDComponents)
	record, err := f.client.FetchClusterRole(ctx, p.IDComponents.UniqueComponent())
	return handleFetchResult(record, err)
}

// cluster role bindings

type clusterRoleBindingNodeResolver struct {
	client RBACClient
}

func registerClusterRoleBindingNodeResolver(register relay.NodeRegister, client RBACClient) {
	resolver := &clusterRoleBindingNodeResolver{client: client}
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.ClusterRoleBindingType,
		Translator: globalid.ClusterRoleBindingTranslator,
		Resolve:    resolver.fetch,
	})
}

func (f *clusterRoleBindingNodeResolver) fetch(p relay.NodeResolverParams) (interface{}, error) {
	ctx := setContextFromComponents(p.Context, p.IDComponents)
	record, err := f.client.FetchClusterRoleBinding(ctx, p.IDComponents.UniqueComponent())
	return handleFetchResult(record, err)
}

// roles

type roleNodeResolver struct {
	client RBACClient
}

func registerRoleNodeResolver(register relay.NodeRegister, client RBACClient) {
	resolver := &roleNodeResolver{client: client}
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.RoleType,
		Translator: globalid.RoleTranslator,
		Resolve:    resolver.fetch,
	})
}

func (f *roleNodeResolver) fetch(p relay.NodeResolverParams) (interface{}, error) {
	ctx := setContextFromComponents(p.Context, p.IDComponents)
	record, err := f.client.FetchRole(ctx, p.IDComponents.UniqueComponent())
	return handleFetchResult(record, err)
}

// role bindings

type roleBindingNodeResolver struct {
	client RBACClient
}

func registerRoleBindingNodeResolver(register relay.NodeRegister, client RBACClient) {
	resolver := &roleBindingNodeResolver{client: client}
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.RoleBindingType,
		Translator: globalid.RoleBindingTranslator,
		Resolve:    resolver.fetch,
	})
}

func (f *roleBindingNodeResolver) fetch(p relay.NodeResolverParams) (interface{}, error) {
	ctx := setContextFromComponents(p.Context, p.IDComponents)
	record, err := f.client.FetchRoleBinding(ctx, p.IDComponents.UniqueComponent())
	return handleFetchResult(record, err)
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

type silencedNodeResolver struct {
	client SilencedClient
}

func registerSilencedNodeResolver(register relay.NodeRegister, client SilencedClient) {
	resolver := &silencedNodeResolver{client: client}
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.SilencedType,
		Translator: globalid.SilenceTranslator,
		Resolve:    resolver.fetch,
	})
}

func (f *silencedNodeResolver) fetch(p relay.NodeResolverParams) (interface{}, error) {
	ctx := setContextFromComponents(p.Context, p.IDComponents)
	record, err := f.client.GetSilencedByName(ctx, p.IDComponents.UniqueComponent())
	return handleFetchResult(record, err)
}

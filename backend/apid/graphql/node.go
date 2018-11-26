package graphql

import (
	"context"
	"errors"

	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/apid/graphql/relay"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/graphql"
)

//
// Node Resolver
//

type nodeResolver struct {
	register relay.NodeRegister
}

func newNodeResolver(factory ClientFactory) *nodeResolver {
	register := relay.NodeRegister{}

	registerAssetNodeResolver(register, factory)
	registerCheckNodeResolver(register, factory)
	registerEntityNodeResolver(register, factory)
	registerHandlerNodeResolver(register, factory)
	registerHookNodeResolver(register, factory)
	registerMutatorNodeResolver(register, factory)
	registerClusterRoleNodeResolver(register, factory)
	registerClusterRoleBindingNodeResolver(register, factory)
	registerRoleNodeResolver(register, factory)
	registerRoleBindingNodeResolver(register, factory)
	registerUserNodeResolver(register, factory)
	registerEventNodeResolver(register, factory)
	registerNamespaceNodeResolver(register, factory)

	return &nodeResolver{register}
}

func (r *nodeResolver) FindType(i interface{}) *graphql.Type {
	translator, err := globalid.ReverseLookup(i)
	if err != nil {
		return nil
	}

	components := translator.Encode(i)
	resolver := r.register.Lookup(components)
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
	factory ClientFactory
}

func registerAssetNodeResolver(register relay.NodeRegister, factory ClientFactory) {
	resolver := &assetNodeResolver{factory}
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.AssetType,
		Translator: globalid.AssetTranslator,
		Resolve:    resolver.fetch,
	})
}

func (f *assetNodeResolver) fetch(p relay.NodeResolverParams) (interface{}, error) {
	ctx := setContextFromComponents(p.Context, p.IDComponents)
	client := f.factory.NewWithContext(ctx)
	record, err := client.FetchAsset(p.IDComponents.UniqueComponent())
	return handleFetchResult(record, err)
}

// checks

type checkNodeResolver struct {
	factory ClientFactory
}

func registerCheckNodeResolver(register relay.NodeRegister, factory ClientFactory) {
	resolver := &checkNodeResolver{factory}
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.CheckConfigType,
		Translator: globalid.CheckTranslator,
		Resolve:    resolver.fetch,
	})
}

func (f *checkNodeResolver) fetch(p relay.NodeResolverParams) (interface{}, error) {
	ctx := setContextFromComponents(p.Context, p.IDComponents)
	client := f.factory.NewWithContext(ctx)
	record, err := client.FetchCheck(p.IDComponents.UniqueComponent())
	return handleFetchResult(record, err)
}

// entities

type entityNodeResolver struct {
	factory ClientFactory
}

func registerEntityNodeResolver(register relay.NodeRegister, factory ClientFactory) {
	resolver := &entityNodeResolver{factory}
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.EntityType,
		Translator: globalid.EntityTranslator,
		Resolve:    resolver.fetch,
	})
}

func (f *entityNodeResolver) fetch(p relay.NodeResolverParams) (interface{}, error) {
	ctx := setContextFromComponents(p.Context, p.IDComponents)
	client := f.factory.NewWithContext(ctx)
	record, err := client.FetchEntity(p.IDComponents.UniqueComponent())
	return handleFetchResult(record, err)
}

// handlers

type handlerNodeResolver struct {
	factory ClientFactory
}

func registerHandlerNodeResolver(register relay.NodeRegister, factory ClientFactory) {
	resolver := &handlerNodeResolver{factory}
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.HandlerType,
		Translator: globalid.HandlerTranslator,
		Resolve:    resolver.fetch,
	})
}

func (f *handlerNodeResolver) fetch(p relay.NodeResolverParams) (interface{}, error) {
	ctx := setContextFromComponents(p.Context, p.IDComponents)
	client := f.factory.NewWithContext(ctx)
	record, err := client.FetchHandler(p.IDComponents.UniqueComponent())
	return handleFetchResult(record, err)
}

// hooks

type hookNodeResolver struct {
	factory ClientFactory
}

func registerHookNodeResolver(register relay.NodeRegister, factory ClientFactory) {
	resolver := &hookNodeResolver{factory}
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.HookConfigType,
		Translator: globalid.HookTranslator,
		Resolve:    resolver.fetch,
	})
}

func (f *hookNodeResolver) fetch(p relay.NodeResolverParams) (interface{}, error) {
	ctx := setContextFromComponents(p.Context, p.IDComponents)
	client := f.factory.NewWithContext(ctx)
	record, err := client.FetchHook(p.IDComponents.UniqueComponent())
	return handleFetchResult(record, err)
}

// mutators

type mutatorNodeResolver struct {
	factory ClientFactory
}

func registerMutatorNodeResolver(register relay.NodeRegister, factory ClientFactory) {
	resolver := &mutatorNodeResolver{factory}
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.MutatorType,
		Translator: globalid.MutatorTranslator,
		Resolve:    resolver.fetch,
	})
}

func (f *mutatorNodeResolver) fetch(p relay.NodeResolverParams) (interface{}, error) {
	ctx := setContextFromComponents(p.Context, p.IDComponents)
	client := f.factory.NewWithContext(ctx)
	record, err := client.FetchMutator(p.IDComponents.UniqueComponent())
	return handleFetchResult(record, err)
}

// cluster roles

type clusterRoleNodeResolver struct {
	factory ClientFactory
}

func registerClusterRoleNodeResolver(register relay.NodeRegister, factory ClientFactory) {
	resolver := &clusterRoleNodeResolver{factory}
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.ClusterRoleType,
		Translator: globalid.ClusterRoleTranslator,
		Resolve:    resolver.fetch,
	})
}

func (f *clusterRoleNodeResolver) fetch(p relay.NodeResolverParams) (interface{}, error) {
	ctx := setContextFromComponents(p.Context, p.IDComponents)
	client := f.factory.NewWithContext(ctx)
	record, err := client.FetchClusterRole(p.IDComponents.UniqueComponent())
	return handleFetchResult(record, err)
}

// cluster role bindings

type clusterRoleBindingNodeResolver struct {
	factory ClientFactory
}

func registerClusterRoleBindingNodeResolver(register relay.NodeRegister, factory ClientFactory) {
	resolver := &clusterRoleBindingNodeResolver{factory}
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.ClusterRoleBindingType,
		Translator: globalid.ClusterRoleBindingTranslator,
		Resolve:    resolver.fetch,
	})
}

func (f *clusterRoleBindingNodeResolver) fetch(p relay.NodeResolverParams) (interface{}, error) {
	ctx := setContextFromComponents(p.Context, p.IDComponents)
	client := f.factory.NewWithContext(ctx)
	record, err := client.FetchClusterRoleBinding(p.IDComponents.UniqueComponent())
	return handleFetchResult(record, err)
}

// roles

type roleNodeResolver struct {
	factory ClientFactory
}

func registerRoleNodeResolver(register relay.NodeRegister, factory ClientFactory) {
	resolver := &roleNodeResolver{factory}
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.RoleType,
		Translator: globalid.RoleTranslator,
		Resolve:    resolver.fetch,
	})
}

func (f *roleNodeResolver) fetch(p relay.NodeResolverParams) (interface{}, error) {
	ctx := setContextFromComponents(p.Context, p.IDComponents)
	client := f.factory.NewWithContext(ctx)
	record, err := client.FetchRole(p.IDComponents.UniqueComponent())
	return handleFetchResult(record, err)
}

// role bindings

type roleBindingNodeResolver struct {
	factory ClientFactory
}

func registerRoleBindingNodeResolver(register relay.NodeRegister, factory ClientFactory) {
	resolver := &roleBindingNodeResolver{factory}
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.RoleBindingType,
		Translator: globalid.RoleBindingTranslator,
		Resolve:    resolver.fetch,
	})
}

func (f *roleBindingNodeResolver) fetch(p relay.NodeResolverParams) (interface{}, error) {
	ctx := setContextFromComponents(p.Context, p.IDComponents)
	client := f.factory.NewWithContext(ctx)
	record, err := client.FetchRoleBinding(p.IDComponents.UniqueComponent())
	return handleFetchResult(record, err)
}

// user

type userNodeResolver struct {
	factory ClientFactory
}

func registerUserNodeResolver(register relay.NodeRegister, factory ClientFactory) {
	resolver := &userNodeResolver{factory}
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.UserType,
		Translator: globalid.UserTranslator,
		Resolve:    resolver.fetch,
	})
}

func (f *userNodeResolver) fetch(p relay.NodeResolverParams) (interface{}, error) {
	ctx := setContextFromComponents(p.Context, p.IDComponents)
	client := f.factory.NewWithContext(ctx)
	record, err := client.FetchUser(p.IDComponents.UniqueComponent())
	return handleFetchResult(record, err)
}

// events

type eventNodeResolver struct {
	factory ClientFactory
}

func registerEventNodeResolver(register relay.NodeRegister, factory ClientFactory) {
	resolver := &eventNodeResolver{factory}
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
	client := f.factory.NewWithContext(ctx)
	record, err := client.FetchEvent(evComponents.EntityName(), evComponents.CheckName())
	return handleFetchResult(record, err)
}

// namespaces

type namespaceNodeResolver struct {
	factory ClientFactory
}

func registerNamespaceNodeResolver(register relay.NodeRegister, factory ClientFactory) {
	resolver := &namespaceNodeResolver{factory}
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.NamespaceType,
		Translator: globalid.NamespaceTranslator,
		Resolve:    resolver.fetch,
	})
}

func (f *namespaceNodeResolver) fetch(p relay.NodeResolverParams) (interface{}, error) {
	client := f.factory.NewWithContext(p.Context)
	record, err := client.FetchNamespace(p.IDComponents.UniqueComponent())
	return handleFetchResult(record, err)
}

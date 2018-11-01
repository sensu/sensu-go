package graphql

import (
	"context"
	"errors"

	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/apid/graphql/relay"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
)

//
// Node Resolver
//

type nodeResolver struct {
	register relay.NodeRegister
}

func newNodeResolver(store store.Store, getter types.QueueGetter) *nodeResolver {
	register := relay.NodeRegister{}

	registerAssetNodeResolver(register, store)
	registerCheckNodeResolver(register, store, getter)
	registerEntityNodeResolver(register, store)
	registerHandlerNodeResolver(register, store)
	registerHookNodeResolver(register, store)
	registerMutatorNodeResolver(register, store)
	registerRoleNodeResolver(register, store)
	registerUserNodeResolver(register, store)
	registerEventNodeResolver(register, store)
	registerNamespaceNodeResolver(register, store)

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
	controller actions.AssetController
}

func registerAssetNodeResolver(register relay.NodeRegister, store store.AssetStore) {
	controller := actions.NewAssetController(store)
	resolver := &assetNodeResolver{controller}
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.AssetType,
		Translator: globalid.AssetTranslator,
		Resolve:    resolver.fetch,
	})
}

func (f *assetNodeResolver) fetch(p relay.NodeResolverParams) (interface{}, error) {
	ctx := setContextFromComponents(p.Context, p.IDComponents)
	record, err := f.controller.Find(ctx, p.IDComponents.UniqueComponent())
	return handleControllerResults(record, err)
}

// checks

type checkNodeResolver struct {
	controller actions.CheckController
}

func registerCheckNodeResolver(register relay.NodeRegister, store store.Store, getter types.QueueGetter) {
	controller := actions.NewCheckController(store, getter)
	resolver := &checkNodeResolver{controller}
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.CheckConfigType,
		Translator: globalid.CheckTranslator,
		Resolve:    resolver.fetch,
	})
}

func (f *checkNodeResolver) fetch(p relay.NodeResolverParams) (interface{}, error) {
	ctx := setContextFromComponents(p.Context, p.IDComponents)
	record, err := f.controller.Find(ctx, p.IDComponents.UniqueComponent())
	return handleControllerResults(record, err)
}

// entities

type entityNodeResolver struct {
	controller actions.EntityController
}

func registerEntityNodeResolver(register relay.NodeRegister, store store.EntityStore) {
	controller := actions.NewEntityController(store)
	resolver := &entityNodeResolver{controller}
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.EntityType,
		Translator: globalid.EntityTranslator,
		Resolve:    resolver.fetch,
	})
}

func (f *entityNodeResolver) fetch(p relay.NodeResolverParams) (interface{}, error) {
	ctx := setContextFromComponents(p.Context, p.IDComponents)
	record, err := f.controller.Find(ctx, p.IDComponents.UniqueComponent())
	return handleControllerResults(record, err)
}

// handlers

type handlerNodeResolver struct {
	controller actions.HandlerController
}

func registerHandlerNodeResolver(register relay.NodeRegister, store store.HandlerStore) {
	controller := actions.NewHandlerController(store)
	resolver := &handlerNodeResolver{controller}
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.HandlerType,
		Translator: globalid.HandlerTranslator,
		Resolve:    resolver.fetch,
	})
}

func (f *handlerNodeResolver) fetch(p relay.NodeResolverParams) (interface{}, error) {
	ctx := setContextFromComponents(p.Context, p.IDComponents)
	record, err := f.controller.Find(ctx, p.IDComponents.UniqueComponent())
	return handleControllerResults(record, err)
}

// hooks

type hookNodeResolver struct {
	controller actions.HookController
}

func registerHookNodeResolver(register relay.NodeRegister, store store.HookConfigStore) {
	controller := actions.NewHookController(store)
	resolver := &hookNodeResolver{controller}
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.HookConfigType,
		Translator: globalid.HookTranslator,
		Resolve:    resolver.fetch,
	})
}

func (f *hookNodeResolver) fetch(p relay.NodeResolverParams) (interface{}, error) {
	ctx := setContextFromComponents(p.Context, p.IDComponents)
	record, err := f.controller.Find(ctx, p.IDComponents.UniqueComponent())
	return handleControllerResults(record, err)
}

// mutators

type mutatorNodeResolver struct {
	controller actions.MutatorController
}

func registerMutatorNodeResolver(register relay.NodeRegister, store store.MutatorStore) {
	controller := actions.NewMutatorController(store)
	resolver := &mutatorNodeResolver{controller}
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.MutatorType,
		Translator: globalid.MutatorTranslator,
		Resolve:    resolver.fetch,
	})
}

func (f *mutatorNodeResolver) fetch(p relay.NodeResolverParams) (interface{}, error) {
	ctx := setContextFromComponents(p.Context, p.IDComponents)
	record, err := f.controller.Find(ctx, p.IDComponents.UniqueComponent())
	return handleControllerResults(record, err)
}

// roles

type roleNodeResolver struct {
	controller actions.RoleController
}

func registerRoleNodeResolver(register relay.NodeRegister, store store.RBACStore) {
	controller := actions.NewRoleController(store)
	resolver := &roleNodeResolver{controller}
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.RoleType,
		Translator: globalid.RoleTranslator,
		Resolve:    resolver.fetch,
	})
}

func (f *roleNodeResolver) fetch(p relay.NodeResolverParams) (interface{}, error) {
	ctx := setContextFromComponents(p.Context, p.IDComponents)
	record, err := f.controller.Find(ctx, p.IDComponents.UniqueComponent())
	return handleControllerResults(record, err)
}

// user

type userNodeResolver struct {
	controller actions.UserController
}

func registerUserNodeResolver(register relay.NodeRegister, store store.Store) {
	controller := actions.NewUserController(store)
	resolver := &userNodeResolver{controller}
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.UserType,
		Translator: globalid.UserTranslator,
		Resolve:    resolver.fetch,
	})
}

func (f *userNodeResolver) fetch(p relay.NodeResolverParams) (interface{}, error) {
	ctx := setContextFromComponents(p.Context, p.IDComponents)
	record, err := f.controller.Find(ctx, p.IDComponents.UniqueComponent())
	return handleControllerResults(record, err)
}

// events

type eventNodeResolver struct {
	controller actions.EventController
}

func registerEventNodeResolver(register relay.NodeRegister, store store.Store) {
	controller := actions.NewEventController(store, nil)
	resolver := &eventNodeResolver{controller}
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
	record, err := f.controller.Find(ctx, evComponents.EntityName(), evComponents.CheckName())
	return handleControllerResults(record, err)
}

// namespaces

type namespaceNodeResolver struct {
	controller actions.NamespacesController
}

func registerNamespaceNodeResolver(register relay.NodeRegister, store store.Store) {
	controller := actions.NewNamespacesController(store)
	resolver := &namespaceNodeResolver{controller}
	register.RegisterResolver(relay.NodeResolver{
		ObjectType: schema.NamespaceType,
		Translator: globalid.NamespaceTranslator,
		Resolve:    resolver.fetch,
	})
}

func (f *namespaceNodeResolver) fetch(p relay.NodeResolverParams) (interface{}, error) {
	record, err := f.controller.Find(p.Context, p.IDComponents.UniqueComponent())
	return handleControllerResults(record, err)
}

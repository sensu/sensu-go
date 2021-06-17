package store

import (
	"context"
	"crypto/tls"
	"sync"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store/patch"
	"github.com/sensu/sensu-go/backend/store/provider"
	"github.com/sensu/sensu-go/types"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var _ Store = new(StoreProxy)

// StoreProxy is a mechanism for providing an EventStore with a replaceable
// underlying implementation. It uses an atomic so that calls are not impeded by
// mutex overhead.
type StoreProxy struct {
	impl Store
	mu   sync.RWMutex
}

func NewStoreProxy(s Store) *StoreProxy {
	return &StoreProxy{
		impl: s,
	}
}

func (s *StoreProxy) do() Store {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.impl
}

// DeleteEventByEntityCheck deletes an event using the given entity and check,
// within the namespace stored in ctx.
func (s *StoreProxy) DeleteEventByEntityCheck(ctx context.Context, entity, check string) error {
	return s.do().DeleteEventByEntityCheck(ctx, entity, check)
}

// GetEvents returns all events in the given ctx's namespace. A nil slice with
// no error is returned if none were found.
func (s *StoreProxy) GetEvents(ctx context.Context, pred *SelectionPredicate) ([]*corev2.Event, error) {
	return s.do().GetEvents(ctx, pred)
}

// GetEventsByEntity returns all events for the given entity within the ctx's
// namespace. A nil slice with no error is returned if none were found.
func (s *StoreProxy) GetEventsByEntity(ctx context.Context, entity string, pred *SelectionPredicate) ([]*corev2.Event, error) {
	return s.do().GetEventsByEntity(ctx, entity, pred)
}

// GetEventByEntityCheck returns an event using the given entity and check,
// within the namespace stored in ctx. The resulting event
// is nil if none was found.
func (s *StoreProxy) GetEventByEntityCheck(ctx context.Context, entity, check string) (*corev2.Event, error) {
	return s.do().GetEventByEntityCheck(ctx, entity, check)
}

// UpdateEvent creates or updates a given event. It returns the updated
// event, which may be the same as the event that was passed in, and the
// previous event, if one existed, as well as any error that occurred.
func (s *StoreProxy) UpdateEvent(ctx context.Context, event *corev2.Event) (old, new *corev2.Event, err error) {
	return s.do().UpdateEvent(ctx, event)
}

// DeleteAssetByName deletes an asset using the given name and the
// namespace stored in ctx.
func (s *StoreProxy) DeleteAssetByName(ctx context.Context, name string) error {
	return s.do().DeleteAssetByName(ctx, name)
}

// GetAssets returns all assets in the given ctx's namespace. A nil
// slice with no error is returned if none were found.
func (s *StoreProxy) GetAssets(ctx context.Context, pred *SelectionPredicate) ([]*corev2.Asset, error) {
	return s.do().GetAssets(ctx, pred)
}

// GetAssetByName returns an asset using the given name and the namespace
// stored in ctx. The resulting asset is nil if none was found.
func (s *StoreProxy) GetAssetByName(ctx context.Context, name string) (*corev2.Asset, error) {
	return s.do().GetAssetByName(ctx, name)
}

// UpdateAsset creates or updates a given asset.
func (s *StoreProxy) UpdateAsset(ctx context.Context, asset *corev2.Asset) error {
	return s.do().UpdateAsset(ctx, asset)
}

// CreateJWTSecret create the given JWT secret and returns an error if it was
// unsuccessful or if a secret already exists.
func (s *StoreProxy) CreateJWTSecret(secret []byte) error {
	return s.do().CreateJWTSecret(secret)
}

// GetJWTSecret returns the JWT secret.
func (s *StoreProxy) GetJWTSecret() ([]byte, error) {
	return s.do().GetJWTSecret()
}

// UpdateJWTSecret updates the JWT secret with the given secret.
func (s *StoreProxy) UpdateJWTSecret(secret []byte) error {
	return s.do().UpdateJWTSecret(secret)
}

// DeleteCheckConfigByName deletes a check's configuration using the given name
// and the namespace stored in ctx.
func (s *StoreProxy) DeleteCheckConfigByName(ctx context.Context, name string) error {
	return s.do().DeleteCheckConfigByName(ctx, name)
}

// GetCheckConfigs returns all checks configurations in the given ctx's
// namespace. A nil slice with no error is returned if none
// were found.
func (s *StoreProxy) GetCheckConfigs(ctx context.Context, pred *SelectionPredicate) ([]*corev2.CheckConfig, error) {
	return s.do().GetCheckConfigs(ctx, pred)
}

// GetCheckConfigByName returns a check's configuration using the given name
// and the namespace stored in ctx. The resulting check is
// nil if none was found.
func (s *StoreProxy) GetCheckConfigByName(ctx context.Context, name string) (*corev2.CheckConfig, error) {
	return s.do().GetCheckConfigByName(ctx, name)
}

// UpdateCheckConfig creates or updates a given check's configuration.
func (s *StoreProxy) UpdateCheckConfig(ctx context.Context, check *corev2.CheckConfig) error {
	return s.do().UpdateCheckConfig(ctx, check)
}

// GetCheckConfigWatcher returns a channel that emits CheckConfigWatchEvents notifying
// the caller that a CheckConfig was updated. If the watcher runs into a terminal error
// or the context passed is cancelled, then the channel will be closed. The caller must
// restart the watcher, if needed.
func (s *StoreProxy) GetCheckConfigWatcher(ctx context.Context) <-chan WatchEventCheckConfig {
	return s.do().GetCheckConfigWatcher(ctx)
}

// CreateClusterID creates a sensu cluster id
func (s *StoreProxy) CreateClusterID(ctx context.Context, id string) error {
	return s.do().CreateClusterID(ctx, id)
}

// GetClusterID gets the sensu cluster id
func (s *StoreProxy) GetClusterID(ctx context.Context) (string, error) {
	return s.do().GetClusterID(ctx)
}
func (s *StoreProxy) CreateClusterRoleBinding(ctx context.Context, clusterRoleBinding *types.ClusterRoleBinding) error {
	return s.do().CreateClusterRoleBinding(ctx, clusterRoleBinding)
}

func (s *StoreProxy) CreateOrUpdateClusterRoleBinding(ctx context.Context, clusterRoleBinding *types.ClusterRoleBinding) error {
	return s.do().CreateOrUpdateClusterRoleBinding(ctx, clusterRoleBinding)
}

func (s *StoreProxy) DeleteClusterRoleBinding(ctx context.Context, name string) error {
	return s.do().DeleteClusterRoleBinding(ctx, name)
}

func (s *StoreProxy) GetClusterRoleBinding(ctx context.Context, name string) (*types.ClusterRoleBinding, error) {
	return s.do().GetClusterRoleBinding(ctx, name)
}

func (s *StoreProxy) ListClusterRoleBindings(ctx context.Context, pred *SelectionPredicate) (clusterRoleBindings []*types.ClusterRoleBinding, err error) {
	return s.do().ListClusterRoleBindings(ctx, pred)
}

func (s *StoreProxy) UpdateClusterRoleBinding(ctx context.Context, clusterRoleBinding *types.ClusterRoleBinding) error {
	return s.do().UpdateClusterRoleBinding(ctx, clusterRoleBinding)
}

func (s *StoreProxy) CreateClusterRole(ctx context.Context, clusterRole *types.ClusterRole) error {
	return s.do().CreateClusterRole(ctx, clusterRole)
}

func (s *StoreProxy) CreateOrUpdateClusterRole(ctx context.Context, clusterRole *types.ClusterRole) error {
	return s.do().CreateOrUpdateClusterRole(ctx, clusterRole)
}

func (s *StoreProxy) DeleteClusterRole(ctx context.Context, name string) error {
	return s.do().DeleteClusterRole(ctx, name)
}

func (s *StoreProxy) GetClusterRole(ctx context.Context, name string) (*types.ClusterRole, error) {
	return s.do().GetClusterRole(ctx, name)
}

func (s *StoreProxy) ListClusterRoles(ctx context.Context, pred *SelectionPredicate) (clusterRoles []*types.ClusterRole, err error) {
	return s.do().ListClusterRoles(ctx, pred)
}

func (s *StoreProxy) UpdateClusterRole(ctx context.Context, clusterRole *types.ClusterRole) error {
	return s.do().UpdateClusterRole(ctx, clusterRole)
}

// DeleteHookConfigByName deletes a hook's configuration using the given name
// and the namespace stored in ctx.
func (s *StoreProxy) DeleteHookConfigByName(ctx context.Context, name string) error {
	return s.do().DeleteHookConfigByName(ctx, name)
}

// GetHookConfigs returns all hooks configurations in the given ctx's
// namespace. A nil slice with no error is returned if none
// were found.
func (s *StoreProxy) GetHookConfigs(ctx context.Context, pred *SelectionPredicate) ([]*types.HookConfig, error) {
	return s.do().GetHookConfigs(ctx, pred)
}

// GetHookConfigByName returns a hook's configuration using the given name and
// the namespace stored in ctx. The resulting hook is nil if none was found.
func (s *StoreProxy) GetHookConfigByName(ctx context.Context, name string) (*types.HookConfig, error) {
	return s.do().GetHookConfigByName(ctx, name)
}

// UpdateHookConfig creates or updates a given hook's configuration.
func (s *StoreProxy) UpdateHookConfig(ctx context.Context, check *types.HookConfig) error {
	return s.do().UpdateHookConfig(ctx, check)
}

// DeleteEntity deletes an entity using the given entity struct.
func (s *StoreProxy) DeleteEntity(ctx context.Context, entity *types.Entity) error {
	return s.do().DeleteEntity(ctx, entity)
}

// DeleteEntityByName deletes an entity using the given name and the
// namespace stored in ctx.
func (s *StoreProxy) DeleteEntityByName(ctx context.Context, name string) error {
	return s.do().DeleteEntityByName(ctx, name)
}

// GetEntities returns all entities in the given ctx's namespace. A nil slice
// with no error is returned if none were found.
func (s *StoreProxy) GetEntities(ctx context.Context, pred *SelectionPredicate) ([]*types.Entity, error) {
	return s.do().GetEntities(ctx, pred)
}

// GetEntityByName returns an entity using the given name and the namespace stored
// in ctx. The resulting entity is nil if none was found.
func (s *StoreProxy) GetEntityByName(ctx context.Context, name string) (*types.Entity, error) {
	return s.do().GetEntityByName(ctx, name)
}

// UpdateEntity creates or updates a given entity.
func (s *StoreProxy) UpdateEntity(ctx context.Context, entity *types.Entity) error {
	return s.do().UpdateEntity(ctx, entity)
}

// DeleteEventFilterByName deletes an event filter using the given name and the
// namespace stored in ctx.
func (s *StoreProxy) DeleteEventFilterByName(ctx context.Context, name string) error {
	return s.do().DeleteEventFilterByName(ctx, name)
}

// GetEventFilters returns all filters in the given ctx's namespace. A nil
// slice with no error is returned if none were found.
func (s *StoreProxy) GetEventFilters(ctx context.Context, pred *SelectionPredicate) ([]*types.EventFilter, error) {
	return s.do().GetEventFilters(ctx, pred)
}

// GetEventFilterByName returns a filter using the given name and the
// namespace stored in ctx. The resulting filter is nil if none was found.
func (s *StoreProxy) GetEventFilterByName(ctx context.Context, name string) (*types.EventFilter, error) {
	return s.do().GetEventFilterByName(ctx, name)
}

// UpdateEventFilter creates or updates a given filter.
func (s *StoreProxy) UpdateEventFilter(ctx context.Context, filter *types.EventFilter) error {
	return s.do().UpdateEventFilter(ctx, filter)
}

// DeleteHandlerByName deletes a handler using the given name and the
// namespace stored in ctx.
func (s *StoreProxy) DeleteHandlerByName(ctx context.Context, name string) error {
	return s.do().DeleteHandlerByName(ctx, name)
}

// GetHandlers returns all handlers in the given ctx's namespace. A nil slice
// with no error is returned if none were found.
func (s *StoreProxy) GetHandlers(ctx context.Context, pred *SelectionPredicate) ([]*types.Handler, error) {
	return s.do().GetHandlers(ctx, pred)
}

// GetHandlerByName returns a handler using the given name and the namespace
// stored in ctx. The resulting handler is nil if none was found.
func (s *StoreProxy) GetHandlerByName(ctx context.Context, name string) (*types.Handler, error) {
	return s.do().GetHandlerByName(ctx, name)
}

// UpdateHandler creates or updates a given handler.
func (s *StoreProxy) UpdateHandler(ctx context.Context, handler *types.Handler) error {
	return s.do().UpdateHandler(ctx, handler)
}
func (s *StoreProxy) GetClusterHealth(ctx context.Context, cluster clientv3.Cluster, etcdClientTLSConfig *tls.Config) *types.HealthResponse {
	return s.do().GetClusterHealth(ctx, cluster, etcdClientTLSConfig)
}

// DeleteFailingKeepalive deletes a failing keepalive record for a given entity.
func (s *StoreProxy) DeleteFailingKeepalive(ctx context.Context, entity *types.Entity) error {
	return s.do().DeleteFailingKeepalive(ctx, entity)
}

// GetFailingKeepalives returns a slice of failing keepalives.
func (s *StoreProxy) GetFailingKeepalives(ctx context.Context) ([]*types.KeepaliveRecord, error) {
	return s.do().GetFailingKeepalives(ctx)
}

// UpdateFailingKeepalive updates the given entity keepalive with the given expiration
// in unix timestamp format
func (s *StoreProxy) UpdateFailingKeepalive(ctx context.Context, entity *types.Entity, expiration int64) error {
	return s.do().UpdateFailingKeepalive(ctx, entity, expiration)
}

// DeleteMutatorByName deletes a mutator using the given name and the
// namespace stored in ctx.
func (s *StoreProxy) DeleteMutatorByName(ctx context.Context, name string) error {
	return s.do().DeleteMutatorByName(ctx, name)
}

// GetMutators returns all mutators in the given ctx's namespace. A nil slice
// with no error is returned if none were found.
func (s *StoreProxy) GetMutators(ctx context.Context, pred *SelectionPredicate) ([]*types.Mutator, error) {
	return s.do().GetMutators(ctx, pred)
}

// GetMutatorByName returns a mutator using the given name and the
// namespace stored in ctx. The resulting mutator is nil if
// none was found.
func (s *StoreProxy) GetMutatorByName(ctx context.Context, name string) (*types.Mutator, error) {
	return s.do().GetMutatorByName(ctx, name)
}

// UpdateMutator creates or updates a given mutator.
func (s *StoreProxy) UpdateMutator(ctx context.Context, mutator *types.Mutator) error {
	return s.do().UpdateMutator(ctx, mutator)
}

// CreateNamespace creates a given namespace
func (s *StoreProxy) CreateNamespace(ctx context.Context, namespace *types.Namespace) error {
	return s.do().CreateNamespace(ctx, namespace)
}

// DeleteNamespace deletes a namespace using the given name.
func (s *StoreProxy) DeleteNamespace(ctx context.Context, name string) error {
	return s.do().DeleteNamespace(ctx, name)
}

// ListNamespaces returns all namespaces. A nil slice with no error is
// returned if none were found.
func (s *StoreProxy) ListNamespaces(ctx context.Context, pred *SelectionPredicate) ([]*types.Namespace, error) {
	return s.do().ListNamespaces(ctx, pred)
}

// GetNamespace returns a namespace using the given name. The
// result is nil if none was found.
func (s *StoreProxy) GetNamespace(ctx context.Context, name string) (*types.Namespace, error) {
	return s.do().GetNamespace(ctx, name)
}

// UpdateNamespace updates an existing namespace.
func (s *StoreProxy) UpdateNamespace(ctx context.Context, org *types.Namespace) error {
	return s.do().UpdateNamespace(ctx, org)
}
func (s *StoreProxy) CreateResource(ctx context.Context, resource corev2.Resource) error {
	return s.do().CreateResource(ctx, resource)
}

func (s *StoreProxy) CreateOrUpdateResource(ctx context.Context, resource corev2.Resource) error {
	return s.do().CreateOrUpdateResource(ctx, resource)
}

func (s *StoreProxy) DeleteResource(ctx context.Context, kind, name string) error {
	return s.do().DeleteResource(ctx, kind, name)
}

func (s *StoreProxy) GetResource(ctx context.Context, name string, resource corev2.Resource) error {
	return s.do().GetResource(ctx, name, resource)
}

func (s *StoreProxy) ListResources(ctx context.Context, kind string, resources interface{}, pred *SelectionPredicate) error {
	return s.do().ListResources(ctx, kind, resources, pred)
}

func (s *StoreProxy) PatchResource(ctx context.Context, resource corev2.Resource, name string, patcher patch.Patcher, condition *ETagCondition) error {
	return s.do().PatchResource(ctx, resource, name, patcher, condition)
}

// Create a given role binding
func (s *StoreProxy) CreateRoleBinding(ctx context.Context, roleBinding *types.RoleBinding) error {
	return s.do().CreateRoleBinding(ctx, roleBinding)
}

// CreateOrUpdateRole overwrites the given role binding
func (s *StoreProxy) CreateOrUpdateRoleBinding(ctx context.Context, roleBinding *types.RoleBinding) error {
	return s.do().CreateOrUpdateRoleBinding(ctx, roleBinding)
}

// DeleteRole deletes a role binding using the given name.
func (s *StoreProxy) DeleteRoleBinding(ctx context.Context, name string) error {
	return s.do().DeleteRoleBinding(ctx, name)
}

// GetRole returns a role binding using the given name. An error is returned
// if no binding was found
func (s *StoreProxy) GetRoleBinding(ctx context.Context, name string) (*types.RoleBinding, error) {
	return s.do().GetRoleBinding(ctx, name)
}

// ListRoles returns all role binding. An error is returned if no binding were
// found
func (s *StoreProxy) ListRoleBindings(ctx context.Context, pred *SelectionPredicate) (roleBindings []*types.RoleBinding, err error) {
	return s.do().ListRoleBindings(ctx, pred)
}

// UpdateRole creates or updates a given role binding.
func (s *StoreProxy) UpdateRoleBinding(ctx context.Context, roleBinding *types.RoleBinding) error {
	return s.do().UpdateRoleBinding(ctx, roleBinding)
}

// Create a given role
func (s *StoreProxy) CreateRole(ctx context.Context, role *types.Role) error {
	return s.do().CreateRole(ctx, role)
}

// CreateOrUpdateRole overwrites the given role
func (s *StoreProxy) CreateOrUpdateRole(ctx context.Context, role *types.Role) error {
	return s.do().CreateOrUpdateRole(ctx, role)
}

// DeleteRole deletes a role using the given name.
func (s *StoreProxy) DeleteRole(ctx context.Context, name string) error {
	return s.do().DeleteRole(ctx, name)
}

// GetRole returns a role using the given name. An error is returned if no
// role was found
func (s *StoreProxy) GetRole(ctx context.Context, name string) (*types.Role, error) {
	return s.do().GetRole(ctx, name)
}

// ListRoles returns all roles. An error is returned if no roles were found
func (s *StoreProxy) ListRoles(ctx context.Context, pred *SelectionPredicate) (roles []*types.Role, err error) {
	return s.do().ListRoles(ctx, pred)
}

// UpdateRole creates or updates a given role.
func (s *StoreProxy) UpdateRole(ctx context.Context, role *types.Role) error {
	return s.do().UpdateRole(ctx, role)
}

// DeleteSilencedEntryByName deletes an entry using the given id.
func (s *StoreProxy) DeleteSilencedEntryByName(ctx context.Context, id ...string) error {
	return s.do().DeleteSilencedEntryByName(ctx, id...)
}

// GetSilencedEntries returns all entries. A nil slice with no error is
// returned if none were found.
func (s *StoreProxy) GetSilencedEntries(ctx context.Context) ([]*types.Silenced, error) {
	return s.do().GetSilencedEntries(ctx)
}

// GetSilencedEntriesByCheckName returns all entries for the given check
// within the ctx's namespace. A nil slice with no error is
// returned if none were found.
func (s *StoreProxy) GetSilencedEntriesByCheckName(ctx context.Context, check string) ([]*types.Silenced, error) {
	return s.do().GetSilencedEntriesByCheckName(ctx, check)
}

// GetSilencedEntriesByCheckName returns all entries for the given subscription
// within the ctx's namespace. A nil slice with no error is
// returned if none were found.
func (s *StoreProxy) GetSilencedEntriesBySubscription(ctx context.Context, subscriptions ...string) ([]*types.Silenced, error) {
	return s.do().GetSilencedEntriesBySubscription(ctx, subscriptions...)
}

// GetSilencedEntryByName returns an entry using the given id and the
// namespace stored in ctx. The resulting entry is nil if
// none was found.
func (s *StoreProxy) GetSilencedEntryByName(ctx context.Context, id string) (*types.Silenced, error) {
	return s.do().GetSilencedEntryByName(ctx, id)
}

// UpdateHandler creates or updates a given entry.
func (s *StoreProxy) UpdateSilencedEntry(ctx context.Context, entry *types.Silenced) error {
	return s.do().UpdateSilencedEntry(ctx, entry)
}

// GetSilencedEntriesByName gets all the named silenced entries.
func (s *StoreProxy) GetSilencedEntriesByName(ctx context.Context, id ...string) ([]*types.Silenced, error) {
	return s.do().GetSilencedEntriesByName(ctx, id...)
}

// CreateOrUpdateTessenConfig creates or updates the tessen configuration
func (s *StoreProxy) CreateOrUpdateTessenConfig(ctx context.Context, cfg *corev2.TessenConfig) error {
	return s.do().CreateOrUpdateTessenConfig(ctx, cfg)
}

// GetTessenConfig gets the tessen configuration
func (s *StoreProxy) GetTessenConfig(ctx context.Context) (*corev2.TessenConfig, error) {
	return s.do().GetTessenConfig(ctx)
}

// GetTessenConfigWatcher returns a tessen config watcher
func (s *StoreProxy) GetTessenConfigWatcher(ctx context.Context) <-chan WatchEventTessenConfig {
	return s.do().GetTessenConfigWatcher(ctx)
}

// AuthenticateUser attempts to authenticate a user with the given username
// and hashed password. An error is returned if the user does not exist, is
// disabled or the given password does not match.
func (s *StoreProxy) AuthenticateUser(ctx context.Context, username, password string) (*types.User, error) {
	return s.do().AuthenticateUser(ctx, username, password)
}

// CreateUsern creates a new user with the given user struct.
func (s *StoreProxy) CreateUser(ctx context.Context, user *types.User) error {
	return s.do().CreateUser(ctx, user)
}

// GetUser returns a user using the given username.
func (s *StoreProxy) GetUser(ctx context.Context, username string) (*types.User, error) {
	return s.do().GetUser(ctx, username)
}

// GetUsers returns all enabled users. A nil slice with no error is
// returned if none were found.
func (s *StoreProxy) GetUsers() ([]*types.User, error) {
	return s.do().GetUsers()
}

// GetUsers returns all users, including the disabled ones. A nil slice with
// no error is  returned if none were found.
func (s *StoreProxy) GetAllUsers(pred *SelectionPredicate) ([]*types.User, error) {
	return s.do().GetAllUsers(pred)
}

// UpdateHandler updates a given user.
func (s *StoreProxy) UpdateUser(user *types.User) error {
	return s.do().UpdateUser(user)
}

// RegisterExtension registers an extension. It associates an extension type and
// name with a URL. The registry assumes that the extension provides
// a handler and a mutator named 'name'.
func (s *StoreProxy) RegisterExtension(ctx context.Context, ext *types.Extension) error {
	return s.do().RegisterExtension(ctx, ext)
}

// DeregisterExtension deregisters an extension. If the extension does not exist,
// nil error is returned.
func (s *StoreProxy) DeregisterExtension(ctx context.Context, name string) error {
	return s.do().DeregisterExtension(ctx, name)
}

// GetExtension gets the address of a registered extension. If the extension does
// not exist, ErrNoExtension is returned.
func (s *StoreProxy) GetExtension(ctx context.Context, name string) (*types.Extension, error) {
	return s.do().GetExtension(ctx, name)
}

// GetExtensions gets all the extensions for the namespace in ctx.
func (s *StoreProxy) GetExtensions(ctx context.Context, pred *SelectionPredicate) ([]*types.Extension, error) {
	return s.do().GetExtensions(ctx, pred)
}

// NewInitializer returns the Initializer interfaces, which provides the
// required mechanism to verify if a store is initialized
func (s *StoreProxy) NewInitializer(ctx context.Context) (Initializer, error) {
	return s.do().NewInitializer(ctx)
}

func (s *StoreProxy) GetProviderInfo() *provider.Info {
	p, ok := s.do().(provider.InfoGetter)
	if ok {
		return p.GetProviderInfo()
	}
	return &provider.Info{
		TypeMeta: corev2.TypeMeta{
			Type:       "etcd",
			APIVersion: "core/v2",
		},
	}
}

type closer interface {
	Close() error
}

func (s *StoreProxy) UpdateStore(to Store) {
	old := s.do()
	defer func() {
		if closer, ok := old.(closer); ok {
			// delay closing the old store for a while so that in-flight requests
			// complete without error. Correctness will suffer under high loads,
			// but we don't want the system to crash unnecessarily.
			<-time.After(time.Minute)
			_ = closer.Close()
		}
	}()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.impl = to
}

func (s *StoreProxy) Close() error {
	store := s.do()
	if c, ok := store.(closer); ok {
		return c.Close()
	}
	return nil
}

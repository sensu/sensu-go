package graphql

import (
	"context"

	dto "github.com/prometheus/client_model/go"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
)

type AssetClient interface {
	ListAssets(context.Context) ([]*corev2.Asset, error)
	FetchAsset(context.Context, string) (*corev2.Asset, error)
	CreateAsset(context.Context, *corev2.Asset) error
	UpdateAsset(context.Context, *corev2.Asset) error
}

type CheckClient interface {
	CreateCheck(context.Context, *corev2.CheckConfig) error
	UpdateCheck(context.Context, *corev2.CheckConfig) error
	DeleteCheck(context.Context, string) error
	ExecuteCheck(context.Context, string, *corev2.AdhocRequest) error
	FetchCheck(context.Context, string) (*corev2.CheckConfig, error)
	ListChecks(context.Context) ([]*corev2.CheckConfig, error)
}

type EntityClient interface {
	DeleteEntity(context.Context, string) error
	CreateEntity(context.Context, *corev2.Entity) error
	UpdateEntity(context.Context, *corev2.Entity) error
	FetchEntity(context.Context, string) (*corev2.Entity, error)
	ListEntities(ctx context.Context) ([]*corev2.Entity, error)
}

type EventClient interface {
	UpdateEvent(ctx context.Context, event *corev2.Event) error
	FetchEvent(ctx context.Context, entity, check string) (*corev2.Event, error)
	DeleteEvent(ctx context.Context, entity, check string) error
	ListEvents(ctx context.Context, pred *store.SelectionPredicate) ([]*corev2.Event, error)
}

type EventFilterClient interface {
	ListEventFilters(ctx context.Context) ([]*corev2.EventFilter, error)
	FetchEventFilter(ctx context.Context, name string) (*corev2.EventFilter, error)
	CreateEventFilter(ctx context.Context, filter *corev2.EventFilter) error
	UpdateEventFilter(ctx context.Context, filter *corev2.EventFilter) error
	DeleteEventFilter(ctx context.Context, name string) error
}

type HandlerClient interface {
	ListHandlers(ctx context.Context) ([]*corev2.Handler, error)
	FetchHandler(ctx context.Context, name string) (*corev2.Handler, error)
	CreateHandler(ctx context.Context, handler *corev2.Handler) error
	UpdateHandler(ctx context.Context, handler *corev2.Handler) error
	DeleteHandler(ctx context.Context, name string) error
}

type MutatorClient interface {
	ListMutators(ctx context.Context) ([]*corev2.Mutator, error)
	FetchMutator(ctx context.Context, name string) (*corev2.Mutator, error)
	CreateMutator(ctx context.Context, mutator *corev2.Mutator) error
	UpdateMutator(ctx context.Context, mutator *corev2.Mutator) error
	DeleteMutator(ctx context.Context, name string) error
}

type SilencedClient interface {
	UpdateSilenced(ctx context.Context, silenced *corev2.Silenced) error
	GetSilencedByName(ctx context.Context, name string) (*corev2.Silenced, error)
	DeleteSilencedByName(ctx context.Context, name string) error
	ListSilenced(ctx context.Context) ([]*corev2.Silenced, error)
	GetSilencedByCheckName(ctx context.Context, check string) ([]*corev2.Silenced, error)
	GetSilencedBySubscription(ctx context.Context, subs ...string) ([]*corev2.Silenced, error)
}

type NamespaceClient interface {
	ListNamespaces(ctx context.Context, pred *store.SelectionPredicate) ([]*corev2.Namespace, error)
	FetchNamespace(ctx context.Context, name string) (*corev2.Namespace, error)
	CreateNamespace(ctx context.Context, namespace *corev2.Namespace) error
	UpdateNamespace(ctx context.Context, namespace *corev2.Namespace) error
}

type HookClient interface {
	ListHookConfigs(ctx context.Context) ([]*corev2.HookConfig, error)
	FetchHookConfig(ctx context.Context, name string) (*corev2.HookConfig, error)
	CreateHookConfig(ctx context.Context, hook *corev2.HookConfig) error
	UpdateHookConfig(ctx context.Context, hook *corev2.HookConfig) error
}

type UserClient interface {
	ListUsers(ctx context.Context) ([]*corev2.User, error)
	FetchUser(ctx context.Context, name string) (*corev2.User, error)
	CreateUser(ctx context.Context, user *corev2.User) error
	UpdateUser(ctx context.Context, user *corev2.User) error
}

type RBACClient interface {
	ListRoleBindings(ctx context.Context) ([]*corev2.RoleBinding, error)
	FetchRoleBinding(ctx context.Context, name string) (*corev2.RoleBinding, error)
	CreateRoleBinding(ctx context.Context, rb *corev2.RoleBinding) error
	UpdateRoleBinding(ctx context.Context, rb *corev2.RoleBinding) error
	ListRoles(ctx context.Context) ([]*corev2.Role, error)
	FetchRole(ctx context.Context, name string) (*corev2.Role, error)
	CreateRole(ctx context.Context, rb *corev2.Role) error
	UpdateRole(ctx context.Context, rb *corev2.Role) error
	ListClusterRoleBindings(ctx context.Context) ([]*corev2.ClusterRoleBinding, error)
	FetchClusterRoleBinding(ctx context.Context, name string) (*corev2.ClusterRoleBinding, error)
	CreateClusterRoleBinding(ctx context.Context, rb *corev2.ClusterRoleBinding) error
	UpdateClusterRoleBinding(ctx context.Context, rb *corev2.ClusterRoleBinding) error
	ListClusterRoles(ctx context.Context) ([]*corev2.ClusterRole, error)
	FetchClusterRole(ctx context.Context, name string) (*corev2.ClusterRole, error)
	CreateClusterRole(ctx context.Context, rb *corev2.ClusterRole) error
	UpdateClusterRole(ctx context.Context, rb *corev2.ClusterRole) error
}

type GenericClient interface {
	SetTypeMeta(meta corev2.TypeMeta) error
	Create(ctx context.Context, value corev2.Resource) error
	Update(ctx context.Context, value corev2.Resource) error
	Delete(ctx context.Context, name string) error
	Get(ctx context.Context, name string, val corev2.Resource) error
	List(ctx context.Context, resources interface{}, pred *store.SelectionPredicate) error
}

type EtcdHealthController interface {
	GetClusterHealth(ctx context.Context) *corev2.HealthResponse
}

type VersionController interface {
	GetVersion(ctx context.Context) *corev2.Version
}

type MetricGatherer interface {
	Gather() ([]*dto.MetricFamily, error)
}

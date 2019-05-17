package client

import (
	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/types"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// ListOptions represents the various options that can be used when listing
// resources.
type ListOptions struct {
	FieldSelector string
	LabelSelector string

	// ContinueToken is the current pagination token.
	ContinueToken string

	// ChunkSize is the number of objects to fetch per page when taking
	// advantage of the API's pagination capabilities. ChunkSize <= 0 means
	// fetch everything all at once; do not use pagination.
	ChunkSize int
}

// APIClient client methods across the Sensu API
type APIClient interface {
	AuthenticationAPIClient
	AssetAPIClient
	CheckAPIClient
	ClusterRoleAPIClient
	ClusterRoleBindingAPIClient
	EntityAPIClient
	EventAPIClient
	ExtensionAPIClient
	FilterAPIClient
	HandlerAPIClient
	HealthAPIClient
	HookAPIClient
	MutatorAPIClient
	NamespaceAPIClient
	RoleAPIClient
	RoleBindingAPIClient
	UserAPIClient
	SilencedAPIClient
	GenericClient
	ClusterMemberClient
	LicenseClient
}

// GenericClient exposes generic resource methods.
type GenericClient interface {
	// Delete deletes the key with the given path
	Delete(path string) error
	// Get retrieves the key at the given path and stores it into obj
	Get(path string, obj interface{}) error
	// List retrieves all keys with the given path prefix and stores them into objs
	List(path string, objs interface{}, options *ListOptions) error
	// Post creates the given obj at the specified path
	Post(path string, obj interface{}) error
	// Put creates the given obj at the specified path
	Put(path string, obj interface{}) error

	// PutResource puts a resource according to its URIPath.
	PutResource(types.Wrapper) error
}

// AuthenticationAPIClient client methods for authenticating
type AuthenticationAPIClient interface {
	CreateAccessToken(url string, userid string, secret string) (*types.Tokens, error)
	TestCreds(userid string, secret string) error
	Logout(token string) error
	RefreshAccessToken(refreshToken string) (*types.Tokens, error)
}

// AssetAPIClient client methods for assets
type AssetAPIClient interface {
	CreateAsset(*types.Asset) error
	UpdateAsset(*types.Asset) error
	FetchAsset(string) (*types.Asset, error)
	ListAssets(string, *ListOptions) ([]types.Asset, error)
}

// CheckAPIClient client methods for checks
type CheckAPIClient interface {
	CreateCheck(*types.CheckConfig) error
	DeleteCheck(string, string) error
	ExecuteCheck(*types.AdhocRequest) error
	FetchCheck(string) (*types.CheckConfig, error)
	ListChecks(string, *ListOptions) ([]types.CheckConfig, error)
	UpdateCheck(*types.CheckConfig) error

	AddCheckHook(check *types.CheckConfig, checkHook *types.HookList) error
	RemoveCheckHook(check *types.CheckConfig, checkHookType string, hookName string) error
}

// ClusterRoleAPIClient client methods for cluster roles
type ClusterRoleAPIClient interface {
	CreateClusterRole(*types.ClusterRole) error
	DeleteClusterRole(string) error
	FetchClusterRole(string) (*types.ClusterRole, error)
	ListClusterRoles(*ListOptions) ([]types.ClusterRole, error)
}

// ClusterRoleBindingAPIClient client methods for cluster role bindings
type ClusterRoleBindingAPIClient interface {
	CreateClusterRoleBinding(*types.ClusterRoleBinding) error
	DeleteClusterRoleBinding(string) error
	FetchClusterRoleBinding(string) (*types.ClusterRoleBinding, error)
	ListClusterRoleBindings(*ListOptions) ([]types.ClusterRoleBinding, error)
}

// EntityAPIClient client methods for entities
type EntityAPIClient interface {
	CreateEntity(entity *types.Entity) error
	DeleteEntity(string, string) error
	FetchEntity(ID string) (*types.Entity, error)
	ListEntities(string, *ListOptions) ([]types.Entity, error)
	UpdateEntity(entity *types.Entity) error
}

// FilterAPIClient client methods for filters
type FilterAPIClient interface {
	CreateFilter(*types.EventFilter) error
	DeleteFilter(string, string) error
	FetchFilter(string) (*types.EventFilter, error)
	ListFilters(string, *ListOptions) ([]types.EventFilter, error)
	UpdateFilter(*types.EventFilter) error
}

// EventAPIClient client methods for events
type EventAPIClient interface {
	FetchEvent(string, string) (*types.Event, error)
	ListEvents(string, *ListOptions) ([]corev2.Event, error)

	// DeleteEvent deletes the event identified by entity, check.
	DeleteEvent(namespace, entity, check string) error
	UpdateEvent(*types.Event) error
	ResolveEvent(*types.Event) error
}

// ExtensionAPIClient client methods for extensions
type ExtensionAPIClient interface {
	ListExtensions(namespace string, options *ListOptions) ([]types.Extension, error)
	RegisterExtension(*types.Extension) error
	DeregisterExtension(name, namespace string) error
}

// HandlerAPIClient client methods for handlers
type HandlerAPIClient interface {
	CreateHandler(*types.Handler) error
	DeleteHandler(string, string) error
	ListHandlers(string, *ListOptions) ([]types.Handler, error)
	FetchHandler(string) (*types.Handler, error)
	UpdateHandler(*types.Handler) error
}

// HealthAPIClient client methods for health api
type HealthAPIClient interface {
	Health() (*types.HealthResponse, error)
}

// HookAPIClient client methods for hooks
type HookAPIClient interface {
	CreateHook(*types.HookConfig) error
	UpdateHook(*types.HookConfig) error
	DeleteHook(string, string) error
	FetchHook(string) (*types.HookConfig, error)
	ListHooks(string, *ListOptions) ([]types.HookConfig, error)
}

// MutatorAPIClient client methods for mutators
type MutatorAPIClient interface {
	CreateMutator(*types.Mutator) error
	ListMutators(string, *ListOptions) ([]types.Mutator, error)
	DeleteMutator(string, string) error
	FetchMutator(string) (*types.Mutator, error)
	UpdateMutator(*types.Mutator) error
}

// NamespaceAPIClient client methods for namespaces
type NamespaceAPIClient interface {
	CreateNamespace(*types.Namespace) error
	UpdateNamespace(*types.Namespace) error
	DeleteNamespace(string) error
	ListNamespaces(*ListOptions) ([]types.Namespace, error)
	FetchNamespace(string) (*types.Namespace, error)
}

// UserAPIClient client methods for users
type UserAPIClient interface {
	AddGroupToUser(string, string) error
	CreateUser(*types.User) error
	DisableUser(string) error
	FetchUser(string) (*types.User, error)
	ListUsers(*ListOptions) ([]types.User, error)
	ReinstateUser(string) error
	RemoveGroupFromUser(string, string) error
	RemoveAllGroupsFromUser(string) error
	SetGroupsForUser(string, []string) error
	UpdatePassword(string, string) error
}

// RoleAPIClient client methods for roles
type RoleAPIClient interface {
	CreateRole(*types.Role) error
	DeleteRole(string, string) error
	FetchRole(string) (*types.Role, error)
	ListRoles(string, *ListOptions) ([]types.Role, error)
}

// RoleBindingAPIClient client methods for role bindings
type RoleBindingAPIClient interface {
	CreateRoleBinding(*types.RoleBinding) error
	DeleteRoleBinding(string, string) error
	FetchRoleBinding(string) (*types.RoleBinding, error)
	ListRoleBindings(string, *ListOptions) ([]types.RoleBinding, error)
}

// SilencedAPIClient client methods for silenced
type SilencedAPIClient interface {
	// CreateSilenced creates a new silenced entry from its input.
	CreateSilenced(*types.Silenced) error

	// DeleteSilenced deletes an existing silenced entry given its ID.
	DeleteSilenced(namespace string, name string) error

	// ListSilenceds lists all silenced entries, optionally constraining by
	// subscription or check.
	ListSilenceds(namespace, subscription, check string, options *ListOptions) ([]types.Silenced, error)

	// FetchSilenced fetches the silenced entry by ID.
	FetchSilenced(id string) (*types.Silenced, error)

	// UpdateSilenced updates an existing silenced entry.
	UpdateSilenced(*types.Silenced) error
}

// ClusterMemberClient specifies client methods for cluster membership management
type ClusterMemberClient interface {
	// MemberList lists cluster members
	MemberList() (*clientv3.MemberListResponse, error)

	// MemberAdd adds a cluster member
	MemberAdd(peerAddrs []string) (*clientv3.MemberAddResponse, error)

	// MemberUpdate updates a cluster member
	MemberUpdate(id uint64, peerAddrs []string) (*clientv3.MemberUpdateResponse, error)

	// MemberRemove removes a cluster member
	MemberRemove(id uint64) (*clientv3.MemberRemoveResponse, error)
}

// LicenseClient specifies the enteprise client methods for license management.
// This is a temporary workaround until
// https://github.com/sensu/sensu-go/issues/1870 is implemented
type LicenseClient interface {
	// FetchLicense fetches the installed license
	FetchLicense() (interface{}, error)

	// UpdateLicense updates the installed enterprise license
	UpdateLicense(license interface{}) error
}

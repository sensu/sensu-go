package client

import (
	"net/http"

	"github.com/coreos/etcd/clientv3"
	"github.com/go-resty/resty/v2"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/types"
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
	APIKeyClient
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

// APIKeyClient exposes client methods for api keys.
type APIKeyClient interface {
	// PostAPIKey creates an api key and returns the location header.
	PostAPIKey(path string, obj interface{}) (string, error)
}

// GenericClient exposes generic resource methods.
type GenericClient interface {
	// Delete deletes the key with the given path
	Delete(path string) error
	// Get retrieves the key at the given path and stores it into obj
	Get(path string, obj interface{}) error
	// List retrieves all keys with the given path prefix and stores them into objs
	List(path string, objs interface{}, options *ListOptions, header *http.Header) error
	// Post creates the given obj at the specified path
	Post(path string, obj interface{}) error
	// Put creates the given obj at the specified path
	Put(path string, obj interface{}) error
	// PostWithResponse creates the given obj at the specified path, returning the response
	PostWithResponse(path string, obj interface{}) (*resty.Response, error)

	// PutResource puts a resource according to its URIPath.
	PutResource(types.Wrapper) error
}

// AuthenticationAPIClient client methods for authenticating
type AuthenticationAPIClient interface {
	CreateAccessToken(url string, userid string, secret string) (*corev2.Tokens, error)
	TestCreds(userid string, secret string) error
	Logout(token string) error
	RefreshAccessToken(tokens *corev2.Tokens) (*corev2.Tokens, error)
}

// AssetAPIClient client methods for assets
type AssetAPIClient interface {
	CreateAsset(*corev2.Asset) error
	UpdateAsset(*corev2.Asset) error
	FetchAsset(string) (*corev2.Asset, error)
}

// CheckAPIClient client methods for checks
type CheckAPIClient interface {
	CreateCheck(*corev2.CheckConfig) error
	DeleteCheck(string, string) error
	ExecuteCheck(*corev2.AdhocRequest) error
	FetchCheck(string) (*corev2.CheckConfig, error)
	UpdateCheck(*corev2.CheckConfig) error

	AddCheckHook(check *corev2.CheckConfig, checkHook *corev2.HookList) error
	RemoveCheckHook(check *corev2.CheckConfig, checkHookType string, hookName string) error
}

// ClusterRoleAPIClient client methods for cluster roles
type ClusterRoleAPIClient interface {
	CreateClusterRole(*corev2.ClusterRole) error
	DeleteClusterRole(string) error
	FetchClusterRole(string) (*corev2.ClusterRole, error)
}

// ClusterRoleBindingAPIClient client methods for cluster role bindings
type ClusterRoleBindingAPIClient interface {
	CreateClusterRoleBinding(*corev2.ClusterRoleBinding) error
	DeleteClusterRoleBinding(string) error
	FetchClusterRoleBinding(string) (*corev2.ClusterRoleBinding, error)
}

// EntityAPIClient client methods for entities
type EntityAPIClient interface {
	CreateEntity(entity *corev2.Entity) error
	DeleteEntity(string, string) error
	FetchEntity(ID string) (*corev2.Entity, error)
	UpdateEntity(entity *corev2.Entity) error
}

// FilterAPIClient client methods for filters
type FilterAPIClient interface {
	CreateFilter(*corev2.EventFilter) error
	DeleteFilter(string, string) error
	FetchFilter(string) (*corev2.EventFilter, error)
	UpdateFilter(*corev2.EventFilter) error
}

// EventAPIClient client methods for events
type EventAPIClient interface {
	FetchEvent(string, string) (*corev2.Event, error)

	// DeleteEvent deletes the event identified by entity, check.
	DeleteEvent(namespace, entity, check string) error
	UpdateEvent(*corev2.Event) error
	ResolveEvent(*corev2.Event) error
}

// ExtensionAPIClient client methods for extensions
type ExtensionAPIClient interface {
	RegisterExtension(*corev2.Extension) error
	DeregisterExtension(name, namespace string) error
}

// HandlerAPIClient client methods for handlers
type HandlerAPIClient interface {
	CreateHandler(*corev2.Handler) error
	DeleteHandler(string, string) error
	FetchHandler(string) (*corev2.Handler, error)
	UpdateHandler(*corev2.Handler) error
}

// HealthAPIClient client methods for health api
type HealthAPIClient interface {
	Health() (*corev2.HealthResponse, error)
}

// HookAPIClient client methods for hooks
type HookAPIClient interface {
	CreateHook(*corev2.HookConfig) error
	UpdateHook(*corev2.HookConfig) error
	DeleteHook(string, string) error
	FetchHook(string) (*corev2.HookConfig, error)
}

// MutatorAPIClient client methods for mutators
type MutatorAPIClient interface {
	CreateMutator(*corev2.Mutator) error
	DeleteMutator(string, string) error
	FetchMutator(string) (*corev2.Mutator, error)
	UpdateMutator(*corev2.Mutator) error
}

// NamespaceAPIClient client methods for namespaces
type NamespaceAPIClient interface {
	CreateNamespace(*corev2.Namespace) error
	UpdateNamespace(*corev2.Namespace) error
	DeleteNamespace(string) error
	FetchNamespace(string) (*corev2.Namespace, error)
}

// UserAPIClient client methods for users
type UserAPIClient interface {
	AddGroupToUser(string, string) error
	CreateUser(*corev2.User) error
	DisableUser(string) error
	FetchUser(string) (*corev2.User, error)
	ReinstateUser(string) error
	RemoveGroupFromUser(string, string) error
	RemoveAllGroupsFromUser(string) error
	SetGroupsForUser(string, []string) error
	UpdatePassword(string, string) error
}

// RoleAPIClient client methods for roles
type RoleAPIClient interface {
	CreateRole(*corev2.Role) error
	DeleteRole(string, string) error
	FetchRole(string) (*corev2.Role, error)
}

// RoleBindingAPIClient client methods for role bindings
type RoleBindingAPIClient interface {
	CreateRoleBinding(*corev2.RoleBinding) error
	DeleteRoleBinding(string, string) error
	FetchRoleBinding(string) (*corev2.RoleBinding, error)
}

// SilencedAPIClient client methods for silenced
type SilencedAPIClient interface {
	// CreateSilenced creates a new silenced entry from its input.
	CreateSilenced(*corev2.Silenced) error

	// DeleteSilenced deletes an existing silenced entry given its ID.
	DeleteSilenced(namespace string, name string) error

	// ListSilenceds lists all silenced entries, optionally constraining by
	// subscription or check.
	ListSilenceds(namespace, subscription, check string, options *ListOptions, header *http.Header) ([]types.Silenced, error)

	// FetchSilenced fetches the silenced entry by ID.
	FetchSilenced(id string) (*corev2.Silenced, error)

	// UpdateSilenced updates an existing silenced entry.
	UpdateSilenced(*corev2.Silenced) error
}

// ClusterMemberClient specifies client methods for cluster membership management.
type ClusterMemberClient interface {
	// MemberList lists cluster members.
	MemberList() (*clientv3.MemberListResponse, error)

	// MemberAdd adds a cluster member.
	MemberAdd(peerAddrs []string) (*clientv3.MemberAddResponse, error)

	// MemberUpdate updates a cluster member.
	MemberUpdate(id uint64, peerAddrs []string) (*clientv3.MemberUpdateResponse, error)

	// MemberRemove removes a cluster member.
	MemberRemove(id uint64) (*clientv3.MemberRemoveResponse, error)

	// FetchClusterID gets the sensu cluster id.
	FetchClusterID() (string, error)
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

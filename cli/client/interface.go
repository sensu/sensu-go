package client

import (
	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/types"
)

// APIClient client methods across the Sensu API
type APIClient interface {
	AuthenticationAPIClient
	AssetAPIClient
	CheckAPIClient
	EntityAPIClient
	EnvironmentAPIClient
	EventAPIClient
	ExtensionAPIClient
	FilterAPIClient
	HandlerAPIClient
	HealthAPIClient
	HookAPIClient
	MutatorAPIClient
	OrganizationAPIClient
	RoleAPIClient
	UserAPIClient
	SilencedAPIClient
	GenericClient
	ClusterMemberClient
	LicenseClient
}

// GenericClient exposes generic resource methods.
type GenericClient interface {
	// PutResource puts a resource according to its URIPath.
	PutResource(types.Resource) error
}

// AuthenticationAPIClient client methods for authenticating
type AuthenticationAPIClient interface {
	CreateAccessToken(url string, userid string, secret string) (*types.Tokens, error)
	Logout(token string) error
	RefreshAccessToken(refreshToken string) (*types.Tokens, error)
}

// AssetAPIClient client methods for assets
type AssetAPIClient interface {
	CreateAsset(*types.Asset) error
	UpdateAsset(*types.Asset) error
	FetchAsset(string) (*types.Asset, error)
	ListAssets(string) ([]types.Asset, error)
}

// CheckAPIClient client methods for checks
type CheckAPIClient interface {
	CreateCheck(*types.CheckConfig) error
	DeleteCheck(*types.CheckConfig) error
	ExecuteCheck(*types.AdhocRequest) error
	FetchCheck(string) (*types.CheckConfig, error)
	ListChecks(string) ([]types.CheckConfig, error)
	UpdateCheck(*types.CheckConfig) error

	AddCheckHook(check *types.CheckConfig, checkHook *types.HookList) error
	RemoveCheckHook(check *types.CheckConfig, checkHookType string, hookName string) error
}

// EntityAPIClient client methods for entities
type EntityAPIClient interface {
	CreateEntity(entity *types.Entity) error
	DeleteEntity(entity *types.Entity) error
	FetchEntity(ID string) (*types.Entity, error)
	ListEntities(string) ([]types.Entity, error)
	UpdateEntity(entity *types.Entity) error
}

// FilterAPIClient client methods for filters
type FilterAPIClient interface {
	CreateFilter(*types.EventFilter) error
	DeleteFilter(*types.EventFilter) error
	FetchFilter(string) (*types.EventFilter, error)
	ListFilters(string) ([]types.EventFilter, error)
	UpdateFilter(*types.EventFilter) error
}

// EnvironmentAPIClient client methods for environments
type EnvironmentAPIClient interface {
	CreateEnvironment(string, *types.Environment) error
	DeleteEnvironment(string, string) error
	ListEnvironments(string) ([]types.Environment, error)
	FetchEnvironment(string) (*types.Environment, error)
	UpdateEnvironment(*types.Environment) error
}

// EventAPIClient client methods for events
type EventAPIClient interface {
	FetchEvent(string, string) (*types.Event, error)
	ListEvents(string) ([]types.Event, error)

	// DeleteEvent deletes the event identified by entity, check.
	DeleteEvent(entity, check string) error
	ResolveEvent(*types.Event) error
}

// ExtensionAPIClient client methods for extensions
type ExtensionAPIClient interface {
	ListExtensions(org string) ([]types.Extension, error)
	RegisterExtension(*types.Extension) error
	DeregisterExtension(name, org string) error
}

// HandlerAPIClient client methods for handlers
type HandlerAPIClient interface {
	CreateHandler(*types.Handler) error
	DeleteHandler(*types.Handler) error
	ListHandlers(string) ([]types.Handler, error)
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
	DeleteHook(*types.HookConfig) error
	FetchHook(string) (*types.HookConfig, error)
	ListHooks(string) ([]types.HookConfig, error)
}

// MutatorAPIClient client methods for mutators
type MutatorAPIClient interface {
	CreateMutator(*types.Mutator) error
	ListMutators(string) ([]types.Mutator, error)
	DeleteMutator(*types.Mutator) error
	FetchMutator(string) (*types.Mutator, error)
	UpdateMutator(*types.Mutator) error
}

// OrganizationAPIClient client methods for organizations
type OrganizationAPIClient interface {
	CreateOrganization(*types.Organization) error
	UpdateOrganization(*types.Organization) error
	DeleteOrganization(string) error
	ListOrganizations() ([]types.Organization, error)
	FetchOrganization(string) (*types.Organization, error)
}

// UserAPIClient client methods for users
type UserAPIClient interface {
	AddRoleToUser(string, string) error
	CreateUser(*types.User) error
	DisableUser(string) error
	ListUsers() ([]types.User, error)
	ReinstateUser(string) error
	RemoveRoleFromUser(string, string) error
	UpdatePassword(string, string) error
}

// RoleAPIClient client methods for role
type RoleAPIClient interface {
	CreateRole(*types.Role) error
	DeleteRole(string) error
	FetchRole(string) (*types.Role, error)
	ListRoles() ([]types.Role, error)

	AddRule(role string, rule *types.Rule) error
	RemoveRule(role string, ruleType string) error
}

// SilencedAPIClient client methods for silenced
type SilencedAPIClient interface {
	// CreateSilenced creates a new silenced entry from its input.
	CreateSilenced(*types.Silenced) error

	// DeleteSilenced deletes an existing silenced entry given its ID.
	DeleteSilenced(id string) error

	// ListSilenceds lists all silenced entries, optionally constraining by
	// subscription or check.
	ListSilenceds(org, subscription, check string) ([]types.Silenced, error)

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

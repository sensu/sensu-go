package client

import (
	"github.com/sensu/sensu-go/types"
)

// APIClient client methods across the Sensu API
type APIClient interface {
	AuthenticationAPIClient
	AssetAPIClient
	CheckAPIClient
	EntityAPIClient
	EventAPIClient
	HandlerAPIClient
	OrganizationAPIClient
	UserAPIClient
	RoleAPIClient
}

// AuthenticationAPIClient client methods for authenticating
type AuthenticationAPIClient interface {
	CreateAccessToken(url string, userid string, secret string) (*types.Tokens, error)
	RefreshAccessToken(refreshToken string) (*types.Tokens, error)
}

// AssetAPIClient client methods for assets
type AssetAPIClient interface {
	ListAssets() ([]types.Asset, error)
	CreateAsset(*types.Asset) error
}

// CheckAPIClient client methods for checks
type CheckAPIClient interface {
	CreateCheck(*types.CheckConfig) error
	DeleteCheck(*types.CheckConfig) error
	FetchCheck(string) (*types.CheckConfig, error)
	ListChecks() ([]types.CheckConfig, error)
}

// EntityAPIClient client methods for entities
type EntityAPIClient interface {
	ListEntities() ([]types.Entity, error)
	FetchEntity(ID string) (types.Entity, error)
	DeleteEntity(entity *types.Entity) error
}

// EventAPIClient client methods for events
type EventAPIClient interface {
	FetchEvent(string, string) (*types.Event, error)
	ListEvents() ([]types.Event, error)
}

// HandlerAPIClient client methods for handlers
type HandlerAPIClient interface {
	ListHandlers() ([]types.Handler, error)
	CreateHandler(*types.Handler) error
	DeleteHandler(*types.Handler) error
}

// OrganizationAPIClient client methods for organizations
type OrganizationAPIClient interface {
	CreateOrganization(*types.Organization) error
	DeleteOrganization(string) error
	ListOrganizations() ([]types.Organization, error)
}

// UserAPIClient client methods for users
type UserAPIClient interface {
	CreateUser(*types.User) error
	DeleteUser(string) error
	ListUsers() ([]types.User, error)
}

// RoleAPIClient client methods for role
type RoleAPIClient interface {
	CreateRole(*types.Role) error
	DeleteRole(string) error
	ListRoles() ([]types.Role, error)
	AddRule(role string, rule *types.Rule) error
	RemoveRule(role string, ruleType string) error
}

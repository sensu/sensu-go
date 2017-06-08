package client

import (
	creds "github.com/sensu/sensu-go/cli/client/credentials"
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
	UserAPIClient
}

// AuthenticationAPIClient client methods for authenticating
type AuthenticationAPIClient interface {
	CreateAccessToken(userid string, secret string) (*creds.AccessToken, error)
	RefreshAccessToken(refreshToken string) (*creds.AccessToken, error)
}

// AssetAPIClient client methods for assets
type AssetAPIClient interface {
	ListAssets() ([]types.Asset, error)
	CreateAsset(*types.Asset) error
}

// CheckAPIClient client methods for checks
type CheckAPIClient interface {
	ListChecks() ([]types.CheckConfig, error)
	CreateCheck(*types.CheckConfig) error
	DeleteCheck(*types.CheckConfig) error
}

// EntityAPIClient client methods for entities
type EntityAPIClient interface {
	ListEntities() ([]types.Entity, error)
	FetchEntity(ID string) (types.Entity, error)
	DeleteEntity(entity *types.Entity) error
}

// EventAPIClient client methods for events
type EventAPIClient interface {
	ListEvents() ([]types.Event, error)
}

// HandlerAPIClient client methods for handlers
type HandlerAPIClient interface {
	ListHandlers() ([]types.Handler, error)
	CreateHandler(*types.Handler) error
	DeleteHandler(*types.Handler) error
}

// UserAPIClient client methods for checks
type UserAPIClient interface {
	CreateUser(*types.User) error
	DeleteUser(string) error
	ListUsers() ([]types.User, error)
}

package graphql

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

// checks

type checkCreator interface {
	Create(context.Context, types.CheckConfig) error
}

type checkDestroyer interface {
	Destroy(context.Context, string) error
}

type checkFinder interface {
	Find(ctx context.Context, name string) (*types.CheckConfig, error)
}

type checkExecutor interface {
	QueueAdhocRequest(context.Context, string, *types.AdhocRequest) error
}

type checkReplacer interface {
	CreateOrReplace(context.Context, types.CheckConfig) error
}

// entities

type entityQuerier interface {
	Query(ctx context.Context) ([]*types.Entity, error)
}

type entityFinder interface {
	Find(ctx context.Context, name string) (*types.Entity, error)
}

type entityDestroyer interface {
	Destroy(ctx context.Context, name string) error
}

// events

type eventDestroyer interface {
	Destroy(ctx context.Context, entity, check string) error
}

type eventFinder interface {
	Find(ctx context.Context, entity, check string) (*types.Event, error)
}

type eventQuerier interface {
	Query(ctx context.Context, entity, check string) ([]*types.Event, error)
}

type eventReplacer interface {
	CreateOrReplace(ctx context.Context, event types.Event) error
}

// environments

type namespaceFinder interface {
	Find(ctx context.Context, name string) (*types.Namespace, error)
}

// silences

type silenceCreator interface {
	Create(context.Context, *types.Silenced) error
}

type silenceDestroyer interface {
	Destroy(context.Context, string) error
}

type silenceQuerier interface {
	Query(context.Context, string, string) ([]*types.Silenced, error)
}

// users

type userFinder interface {
	Find(ctx context.Context, name string) (*types.User, error)
}

package graphql

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

// entities

type entityQuerier interface {
	Query(ctx context.Context) ([]*types.Entity, error)
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

type environmentFinder interface {
	Find(ctx context.Context, org, env string) (*types.Environment, error)
}

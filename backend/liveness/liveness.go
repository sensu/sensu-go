package liveness

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
)

// SwitchPrefix contains the base path for switchset, which are tracked under
// path.Join(SwitchPrefix, toggleName, key)
var SwitchPrefix = "/sensu.io/switchsets"

// State represents a custom int type for the key stae
type State int

const (
	// FallbackTTL represents the minimal supported etcd lease TTL,  in case the
	// system encounters a toggle that does not store a TTL
	FallbackTTL = 5

	// Alive state is 0
	Alive State = 0

	// Dead state is 1
	Dead State = 1

	// If a key is marked as buried, it is slated to be deleted
	buried = "buried"
)

func (s State) String() string {
	switch s {
	case Alive:
		return "alive"
	case Dead:
		return "dead"
	default:
		return fmt.Sprintf("invalid<%d>", s)
	}
}

// Interface specifies the interface for liveness
type Interface interface {
	// Alive is an assertion that an entity is alive.
	Alive(ctx context.Context, id string, ttl int64) error

	// Dead is an assertion that an entity is dead. Dead is useful for
	// registering entities that are known to be dead, but not yet tracked.
	Dead(ctx context.Context, id string, ttl int64) error

	// Bury forgets an entity exists
	Bury(ctx context.Context, id string) error

	// BuryAndRevokeLease forgets an entity exists and revokes the lease associated with it
	BuryAndRevokeLease(ctx context.Context, id string) error
}

// Factory is a function that can deliver an Interface
type Factory func(name string, dead, alive EventFunc, logger logrus.FieldLogger) Interface

// EventFunc is a function that can be used by a SwitchSet to handle events.
// The previous state of the switch will be passed to the function.
//
// For "dead" EventFuncs, the leader flag can be used to determine if the
// client that flipped the switch is our client. For "alive" EventFuncs,
// this parameter is always false.
//
// The EventFunc should return whether or not to bury the switch. If bury is
// true, then the key associated with the EventFunc will be buried and no
// further events will occur for this key.
type EventFunc func(key string, prev State, leader bool) (bury bool)

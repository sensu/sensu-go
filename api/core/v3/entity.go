package v3

import (
	"errors"
	fmt "fmt"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// interface assertions guarantee that EntityConfig and EntityState are valid
// Resources.
var (
	_ Resource = new(EntityConfig)
	_ Resource = new(EntityState)
)

type (
	Deregistration   = corev2.Deregistration
	System           = corev2.System
	Network          = corev2.Network
	NetworkInterface = corev2.NetworkInterface
	Process          = corev2.Process
)

// V2EntityToV3 converts a corev2.Entity to an EntityConfig and EntityState.
// The resulting values will contain pointers to e's memory.
func V2EntityToV3(e *corev2.Entity) (*EntityConfig, *EntityState) {
	cfg := EntityConfig{
		Metadata:          &e.ObjectMeta,
		EntityClass:       e.EntityClass,
		User:              e.User,
		Subscriptions:     e.Subscriptions,
		Deregister:        e.Deregister,
		Deregistration:    e.Deregistration,
		KeepaliveHandlers: e.KeepaliveHandlers,
		Redact:            e.Redact,
	}
	state := EntityState{
		Metadata:          &e.ObjectMeta,
		System:            e.System,
		LastSeen:          e.LastSeen,
		SensuAgentVersion: e.SensuAgentVersion,
	}
	return &cfg, &state
}

// V3EntityToV2 converts an EntityConfig and an EntityState to a corev2.Entity.
// Errors are returned if cfg and state's Metadata are nil or not equal in terms
// of their namespace and name. Labels and annotations will be merged, with the
// labels of cfg taking precedence. The resulting object will contain pointers
// to cfg's and state's memory.
func V3EntityToV2(cfg *EntityConfig, state *EntityState) (*corev2.Entity, error) {
	if cfg.Metadata == nil {
		return nil, errors.New("nil EntityConfig metadata")
	}
	if state.Metadata == nil {
		return nil, errors.New("nil EntityState metadata")
	}
	if p, q := cfg.Metadata.Namespace, state.Metadata.Namespace; p != q {
		return nil, fmt.Errorf("EntityConfig.Namespace %s != EntityState.Namespace %s", p, q)
	}
	if p, q := state.Metadata.Name, cfg.Metadata.Name; p != q {
		return nil, fmt.Errorf("EntityConfig.Name %s != EntityState.Name %s", p, q)
	}
	meta := corev2.ObjectMeta{
		Namespace:   cfg.Metadata.Namespace,
		Name:        cfg.Metadata.Name,
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
	}
	for k, v := range state.Metadata.Labels {
		meta.Labels[k] = v
	}
	for k, v := range state.Metadata.Annotations {
		meta.Annotations[k] = v
	}
	for k, v := range cfg.Metadata.Labels {
		meta.Labels[k] = v
	}
	for k, v := range cfg.Metadata.Annotations {
		meta.Annotations[k] = v
	}
	entity := &corev2.Entity{
		ObjectMeta:        meta,
		EntityClass:       cfg.EntityClass,
		User:              cfg.User,
		Subscriptions:     cfg.Subscriptions,
		Deregister:        cfg.Deregister,
		Deregistration:    cfg.Deregistration,
		KeepaliveHandlers: cfg.KeepaliveHandlers,
		Redact:            cfg.Redact,
		System:            state.System,
		LastSeen:          state.LastSeen,
		SensuAgentVersion: state.SensuAgentVersion,
	}
	return entity, nil
}

func FixtureEntityConfig(name string) *EntityConfig {
	return &EntityConfig{
		Metadata: &corev2.ObjectMeta{
			Namespace:   "default",
			Name:        name,
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
		},
		EntityClass:   corev2.EntityAgentClass,
		User:          "agent1",
		Subscriptions: []string{"linux", corev2.GetEntitySubscription(name)},
		Deregister:    true,
		Deregistration: Deregistration{
			Handler: "foo",
		},
		KeepaliveHandlers: []string{
			"alert",
		},
		Redact: []string{
			"password",
		},
	}
}

func FixtureEntityState(name string) *EntityState {
	return &EntityState{
		Metadata: &corev2.ObjectMeta{
			Namespace:   "default",
			Name:        name,
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
		},
		System: System{
			Arch:           "amd64",
			OS:             "linux",
			Platform:       "Gentoo",
			PlatformFamily: "lol",
			Network: Network{
				Interfaces: []NetworkInterface{
					{
						Name: "eth0",
						MAC:  "return of the",
						Addresses: []string{
							"127.0.0.1",
						},
					},
				},
			},
			LibCType:      "glibc",
			VMSystem:      "kvm",
			VMRole:        "host",
			CloudProvider: "aws",
			FloatType:     "hard",
			Processes: []*Process{
				{
					Name: "sensu-agent",
				},
			},
		},
		LastSeen:          12345,
		SensuAgentVersion: "0.0.1",
	}
}

// validate ensures that the entity's class is valid, and that the entity
// is within a namespace.
func (e *EntityConfig) validate() error {
	if err := corev2.ValidateName(e.EntityClass); err != nil {
		return errors.New("entity class " + err.Error())
	}

	if e.Metadata == nil || e.Metadata.Namespace == "" {
		return errors.New("namespace must be set")
	}

	return nil
}

// validate ensures that the entity is within a namespace.
func (e *EntityState) validate() error {
	if e.Metadata == nil || e.Metadata.Namespace == "" {
		return errors.New("namespace must be set")
	}

	return nil
}

// NewEntityState creates a new EntityState.
func NewEntityState(namespace, name string) *EntityState {
	return &EntityState{
		Metadata: &corev2.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
		},
	}
}

// NewEntityConfig creates a new EntityConfig.
func NewEntityConfig(namespace, name string) *EntityConfig {
	return &EntityConfig{
		Metadata: &corev2.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
		},
	}
}

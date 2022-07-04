package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

// EntityStateWrapper is an implementation of storev2.Wrapper, for dealing with
// postgresql entity state storage.
type EntityStateWrapper struct {
	ID                int64
	NamespaceID       int64
	Namespace         string
	EntityConfigID    int64
	Name              string
	LastSeen          int64
	Selectors         []byte
	Annotations       []byte
	Hostname          string
	OS                string
	Platform          string
	PlatformFamily    string
	PlatformVersion   string
	Arch              string
	ARMVersion        int32
	LibCType          string
	VMSystem          string
	VMRole            string
	CloudProvider     string
	FloatType         string
	SensuAgentVersion string
	NetworkNames      []string
	NetworkMACs       []string
	NetworkAddresses  []string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         sql.NullTime
}

// GetName returns the name of the entity.
func (e *EntityStateWrapper) GetName() string {
	return e.Name
}

// GetCreatedAt returns the value of the CreatedAt field
func (e *EntityStateWrapper) GetCreatedAt() time.Time {
	return e.CreatedAt
}

// GetUpdatedAt returns the value of the UpdatedAt field
func (e *EntityStateWrapper) GetUpdatedAt() time.Time {
	return e.UpdatedAt
}

// GetDeletedAt returns the value of the DeletedAt field
func (e *EntityStateWrapper) GetDeletedAt() sql.NullTime {
	return e.DeletedAt
}

// SetCreatedAt sets the value of the CreatedAt field
func (e *EntityStateWrapper) SetCreatedAt(t time.Time) {
	e.CreatedAt = t
}

// SetUpdatedAt sets the value of the UpdatedAt field
func (e *EntityStateWrapper) SetUpdatedAt(t time.Time) {
	e.UpdatedAt = t
}

// SetDeletedAt sets the value of the DeletedAt field
func (e *EntityStateWrapper) SetDeletedAt(t sql.NullTime) {
	e.DeletedAt = t
}

// Unwrap unwraps the EntityStateWrapper into an *EntityState.
func (e *EntityStateWrapper) Unwrap() (corev3.Resource, error) {
	state := new(corev3.EntityState)
	return state, e.unwrapIntoEntityState(state)
}

// WrapEntityState wraps an EntityState into an EntityStateWrapper.
func WrapEntityState(state *corev3.EntityState) storev2.Wrapper {
	meta := state.GetMetadata()
	sys := state.System
	nics := sys.Network.Interfaces
	annotations, _ := json.Marshal(meta.Annotations)
	selectorMap := make(map[string]string)
	for k, v := range meta.Labels {
		k = fmt.Sprintf("labels.%s", k)
		selectorMap[k] = v
	}
	selectors, _ := json.Marshal(selectorMap)
	wrapper := &EntityStateWrapper{
		Namespace:         meta.Namespace,
		Name:              meta.Name,
		Selectors:         selectors,
		Annotations:       annotations,
		LastSeen:          state.LastSeen,
		Hostname:          sys.Hostname,
		OS:                sys.OS,
		Platform:          sys.Platform,
		PlatformFamily:    sys.PlatformFamily,
		PlatformVersion:   sys.PlatformVersion,
		Arch:              sys.Arch,
		ARMVersion:        sys.ARMVersion,
		LibCType:          sys.LibCType,
		VMSystem:          sys.VMSystem,
		VMRole:            sys.VMRole,
		CloudProvider:     sys.CloudProvider,
		FloatType:         sys.FloatType,
		SensuAgentVersion: state.SensuAgentVersion,
	}
	for _, nic := range nics {
		wrapper.NetworkNames = append(wrapper.NetworkNames, nic.Name)
		wrapper.NetworkMACs = append(wrapper.NetworkMACs, nic.MAC)
		addresses, _ := json.Marshal(nic.Addresses)
		wrapper.NetworkAddresses = append(wrapper.NetworkAddresses, string(addresses))
	}
	return wrapper
}

func (e *EntityStateWrapper) unwrapIntoEntityState(state *corev3.EntityState) error {
	state.Metadata = &corev2.ObjectMeta{
		Namespace:   e.Namespace,
		Name:        e.Name,
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
	}
	state.LastSeen = e.LastSeen
	selectors := make(map[string]string)
	if err := json.Unmarshal(e.Selectors, &selectors); err != nil {
		return fmt.Errorf("error unwrapping entity state: %s", err)
	}
	if err := json.Unmarshal(e.Annotations, &state.Metadata.Annotations); err != nil {
		return fmt.Errorf("error unwrapping entity state: %s", err)
	}
	for k, v := range selectors {
		if strings.HasPrefix(k, "labels.") {
			k = strings.TrimPrefix(k, "labels.")
			state.Metadata.Labels[k] = v
		}
	}
	state.SensuAgentVersion = e.SensuAgentVersion
	state.System = corev2.System{
		Hostname:        e.Hostname,
		OS:              e.OS,
		Platform:        e.Platform,
		PlatformFamily:  e.PlatformFamily,
		PlatformVersion: e.PlatformVersion,
		Arch:            e.Arch,
		ARMVersion:      e.ARMVersion,
		LibCType:        e.LibCType,
		VMSystem:        e.VMSystem,
		VMRole:          e.VMRole,
		CloudProvider:   e.CloudProvider,
		FloatType:       e.FloatType,
	}
	if len(e.NetworkNames) != len(e.NetworkMACs) || len(e.NetworkNames) != len(e.NetworkAddresses) {
		return fmt.Errorf("error unwrapping entity state: corrupted entity %s.%s", e.Namespace, e.Name)
	}
	for i := range e.NetworkNames {
		var addresses []string
		if err := json.Unmarshal([]byte(e.NetworkAddresses[i]), &addresses); err != nil {
			return fmt.Errorf("error unwrapping entity state: corrupted entity %s.%s", e.Namespace, e.Name)
		}
		nic := corev2.NetworkInterface{
			Name:      e.NetworkNames[i],
			MAC:       e.NetworkMACs[i],
			Addresses: addresses,
		}
		state.System.Network.Interfaces = append(
			state.System.Network.Interfaces,
			nic,
		)
	}
	return nil
}

// UnwrapInto unwraps an EntityStateWrapper into a provided *EntityState.
// Other data types are not supported.
func (e *EntityStateWrapper) UnwrapInto(face interface{}) error {
	switch state := face.(type) {
	case *corev3.EntityState:
		return e.unwrapIntoEntityState(state)
	default:
		return fmt.Errorf("error unwrapping entity state: unsupported type: %T", state)
	}
}

// SQLParams serializes an EntityStateWrapper into a slice of parameters,
// suitable for passing to a postgresql query.
func (e *EntityStateWrapper) SQLParams() []interface{} {
	return []interface{}{
		&e.Namespace,
		&e.Name,
		&e.LastSeen,
		&e.Selectors,
		&e.Annotations,
		&e.Hostname,
		&e.OS,
		&e.Platform,
		&e.PlatformFamily,
		&e.PlatformVersion,
		&e.Arch,
		&e.ARMVersion,
		&e.LibCType,
		&e.VMSystem,
		&e.VMRole,
		&e.CloudProvider,
		&e.FloatType,
		&e.SensuAgentVersion,
		&e.NetworkNames,
		&e.NetworkMACs,
		&e.NetworkAddresses,
		&e.ID,
		&e.NamespaceID,
		&e.EntityConfigID,
		&e.CreatedAt,
		&e.UpdatedAt,
		&e.DeletedAt,
	}
}

func (e *EntityStateWrapper) TableName() string {
	return new(corev3.EntityState).StoreName()
}

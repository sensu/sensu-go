package postgres

import (
	"encoding/json"
	"fmt"
	"strings"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

// EntityWrapper is an implementation of storev2.Wrapper, for dealing with
// postgresql entity storage.
type EntityWrapper struct {
	Namespace         string
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
}

// Unwrap unwraps the EntityWrapper into an *EntityState.
func (e *EntityWrapper) Unwrap() (corev3.Resource, error) {
	entity := new(corev3.EntityState)
	return entity, e.unwrapIntoEntityState(entity)
}

// WrapEntityState wraps an EntityState into an EntityWrapper.
func WrapEntityState(entity *corev3.EntityState) storev2.Wrapper {
	meta := entity.GetMetadata()
	sys := entity.System
	nics := sys.Network.Interfaces
	annotations, _ := json.Marshal(meta.Annotations)
	selectorMap := make(map[string]string)
	for k, v := range meta.Labels {
		k = fmt.Sprintf("labels.%s", k)
		selectorMap[k] = v
	}
	selectors, _ := json.Marshal(selectorMap)
	wrapper := &EntityWrapper{
		Namespace:         meta.Namespace,
		Name:              meta.Name,
		Selectors:         selectors,
		Annotations:       annotations,
		LastSeen:          entity.LastSeen,
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
		SensuAgentVersion: entity.SensuAgentVersion,
	}
	for _, nic := range nics {
		wrapper.NetworkNames = append(wrapper.NetworkNames, nic.Name)
		wrapper.NetworkMACs = append(wrapper.NetworkMACs, nic.MAC)
		addresses, _ := json.Marshal(nic.Addresses)
		wrapper.NetworkAddresses = append(wrapper.NetworkAddresses, string(addresses))
	}
	return wrapper
}

func (e *EntityWrapper) unwrapIntoEntityState(entity *corev3.EntityState) error {
	entity.Metadata = &corev2.ObjectMeta{
		Namespace:   e.Namespace,
		Name:        e.Name,
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
	}
	entity.LastSeen = e.LastSeen
	selectors := make(map[string]string)
	if err := json.Unmarshal(e.Selectors, &selectors); err != nil {
		return fmt.Errorf("error unwrapping entity state: %s", err)
	}
	if err := json.Unmarshal(e.Annotations, &entity.Metadata.Annotations); err != nil {
		return fmt.Errorf("error unwrapping entity state: %s", err)
	}
	for k, v := range selectors {
		if strings.HasPrefix(k, "labels.") {
			k = strings.TrimPrefix(k, "labels.")
			entity.Metadata.Labels[k] = v
		}
	}
	entity.SensuAgentVersion = e.SensuAgentVersion
	entity.System = corev2.System{
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
		entity.System.Network.Interfaces = append(
			entity.System.Network.Interfaces,
			nic,
		)
	}
	return nil
}

// UnwrapInto unwraps an EntityWrapper into a provided *EntityState. Other data
// types are not supported.
func (e *EntityWrapper) UnwrapInto(face interface{}) error {
	switch entity := face.(type) {
	case *corev3.EntityState:
		return e.unwrapIntoEntityState(entity)
	default:
		return fmt.Errorf("error unwrapping entity state: unsupported type: %T", entity)
	}
}

// SQLParams serializes an EntityWrapper into a slice of parameters, suitable
// for passing to a postgresql query.
func (e *EntityWrapper) SQLParams() []interface{} {
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
	}
}

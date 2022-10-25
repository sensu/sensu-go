package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

// EntityConfigWrapper is an implementation of storev2.Wrapper, for dealing with
// postgresql entity config storage.
type EntityConfigWrapper struct {
	Namespace         string
	Name              string
	Selectors         []byte
	Annotations       []byte
	CreatedBy         string
	EntityClass       string
	User              string
	Subscriptions     []string
	Deregister        bool
	Deregistration    string
	KeepaliveHandlers []string
	Redact            []string
	ID                int64
	NamespaceID       int64
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         sql.NullTime
}

// GetName returns the name of the entity.
func (e *EntityConfigWrapper) GetName() string {
	return e.Name
}

// GetCreatedAt returns the value of the CreatedAt field
func (e *EntityConfigWrapper) GetCreatedAt() time.Time {
	return e.CreatedAt
}

// GetUpdatedAt returns the value of the UpdatedAt field
func (e *EntityConfigWrapper) GetUpdatedAt() time.Time {
	return e.UpdatedAt
}

// GetDeletedAt returns the value of the DeletedAt field
func (e *EntityConfigWrapper) GetDeletedAt() sql.NullTime {
	return e.DeletedAt
}

// SetCreatedAt sets the value of the CreatedAt field
func (e *EntityConfigWrapper) SetCreatedAt(t time.Time) {
	e.CreatedAt = t
}

// SetUpdatedAt sets the value of the UpdatedAt field
func (e *EntityConfigWrapper) SetUpdatedAt(t time.Time) {
	e.UpdatedAt = t
}

// SetDeletedAt sets the value of the DeletedAt field
func (e *EntityConfigWrapper) SetDeletedAt(t sql.NullTime) {
	e.DeletedAt = t
}

// Unwrap unwraps the EntityConfigWrapper into an *EntityConfig.
func (e *EntityConfigWrapper) Unwrap() (corev3.Resource, error) {
	cfg := new(corev3.EntityConfig)
	return cfg, e.unwrapIntoEntityConfig(cfg)
}

// WrapEntityConfig wraps an EntityConfig into an EntityConfigWrapper.
func WrapEntityConfig(cfg *corev3.EntityConfig) storev2.Wrapper {
	meta := cfg.GetMetadata()
	annotations, _ := json.Marshal(meta.Annotations)
	selectorMap := make(map[string]string)
	for k, v := range meta.Labels {
		k = fmt.Sprintf("labels.%s", k)
		selectorMap[k] = v
	}
	selectors, _ := json.Marshal(selectorMap)
	wrapper := &EntityConfigWrapper{
		Namespace:         meta.Namespace,
		Name:              meta.Name,
		Selectors:         selectors,
		Annotations:       annotations,
		CreatedBy:         meta.CreatedBy,
		EntityClass:       cfg.EntityClass,
		User:              cfg.User,
		Subscriptions:     cfg.Subscriptions,
		Deregister:        cfg.Deregister,
		Deregistration:    cfg.Deregistration.Handler,
		KeepaliveHandlers: cfg.KeepaliveHandlers,
		Redact:            cfg.Redact,
	}
	return wrapper
}

func (e *EntityConfigWrapper) unwrapIntoEntityConfig(cfg *corev3.EntityConfig) error {
	cfg.Metadata = &corev2.ObjectMeta{
		Namespace:   e.Namespace,
		Name:        e.Name,
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
		CreatedBy:   e.CreatedBy,
	}
	selectors := make(map[string]string)
	if err := json.Unmarshal(e.Selectors, &selectors); err != nil {
		return fmt.Errorf("error unwrapping entity config: %s", err)
	}
	if err := json.Unmarshal(e.Annotations, &cfg.Metadata.Annotations); err != nil {
		return fmt.Errorf("error unwrapping entity config: %s", err)
	}
	for k, v := range selectors {
		if strings.HasPrefix(k, "labels.") {
			k = strings.TrimPrefix(k, "labels.")
			cfg.Metadata.Labels[k] = v
		}
	}
	cfg.EntityClass = e.EntityClass
	cfg.User = e.User
	cfg.Subscriptions = e.Subscriptions
	cfg.Deregister = e.Deregister
	cfg.Deregistration.Handler = e.Deregistration
	cfg.KeepaliveHandlers = e.KeepaliveHandlers
	cfg.Redact = e.Redact
	return nil
}

// UnwrapInto unwraps an EntityConfigWrapper into a provided *EntityConfig.
// Other data types are not supported.
func (e *EntityConfigWrapper) UnwrapInto(face interface{}) error {
	switch cfg := face.(type) {
	case *corev3.EntityConfig:
		return e.unwrapIntoEntityConfig(cfg)
	default:
		return fmt.Errorf("error unwrapping entity config: unsupported type: %T", cfg)
	}
}

// SQLParams serializes an EntityConfigWrapper into a slice of parameters,
// suitable for passing to a postgresql query.
func (e *EntityConfigWrapper) SQLParams() []interface{} {
	return []interface{}{
		&e.Namespace,
		&e.Name,
		&e.Selectors,
		&e.Annotations,
		&e.CreatedBy,
		&e.EntityClass,
		&e.User,
		&e.Subscriptions,
		&e.Deregister,
		&e.Deregistration,
		&e.KeepaliveHandlers,
		&e.Redact,
		&e.ID,
		&e.NamespaceID,
		&e.CreatedAt,
		&e.UpdatedAt,
		&e.DeletedAt,
	}
}

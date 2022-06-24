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

// NamespaceWrapper is an implementation of storev2.Wrapper, for dealing with
// postgresql namespace storage.
type NamespaceWrapper struct {
	ID          int64
	Name        string
	Selectors   []byte
	Annotations []byte
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   sql.NullTime
}

// GetName returns the name of the namespace.
func (e *NamespaceWrapper) GetName() string {
	return e.Name
}

// GetCreatedAt returns the value of the CreatedAt field
func (e *NamespaceWrapper) GetCreatedAt() time.Time {
	return e.CreatedAt
}

// GetUpdatedAt returns the value of the UpdatedAt field
func (e *NamespaceWrapper) GetUpdatedAt() time.Time {
	return e.UpdatedAt
}

// GetDeletedAt returns the value of the DeletedAt field
func (e *NamespaceWrapper) GetDeletedAt() sql.NullTime {
	return e.DeletedAt
}

// SetCreatedAt sets the value of the CreatedAt field
func (e *NamespaceWrapper) SetCreatedAt(t time.Time) {
	e.CreatedAt = t
}

// SetUpdatedAt sets the value of the UpdatedAt field
func (e *NamespaceWrapper) SetUpdatedAt(t time.Time) {
	e.UpdatedAt = t
}

// SetDeletedAt sets the value of the DeletedAt field
func (e *NamespaceWrapper) SetDeletedAt(t sql.NullTime) {
	e.DeletedAt = t
}

// Unwrap unwraps the NamespaceWrapper into a *Namespace.
func (e *NamespaceWrapper) Unwrap() (corev3.Resource, error) {
	namespace := new(corev3.Namespace)
	return namespace, e.unwrapIntoNamespace(namespace)
}

// WrapNamespace wraps a Namespace into a NamespaceWrapper.
func WrapNamespace(namespace *corev3.Namespace) storev2.Wrapper {
	meta := namespace.GetMetadata()
	annotations, _ := json.Marshal(meta.Annotations)
	selectorMap := make(map[string]string)
	for k, v := range meta.Labels {
		k = fmt.Sprintf("labels.%s", k)
		selectorMap[k] = v
	}
	selectors, _ := json.Marshal(selectorMap)
	wrapper := &NamespaceWrapper{
		Name:        meta.Name,
		Selectors:   selectors,
		Annotations: annotations,
	}
	return wrapper
}

func (e *NamespaceWrapper) unwrapIntoNamespace(namespace *corev3.Namespace) error {
	namespace.Metadata = &corev2.ObjectMeta{
		Name:        e.Name,
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
	}
	selectors := make(map[string]string)
	if err := json.Unmarshal(e.Selectors, &selectors); err != nil {
		return fmt.Errorf("error unwrapping namespace: %s", err)
	}
	if err := json.Unmarshal(e.Annotations, &namespace.Metadata.Annotations); err != nil {
		return fmt.Errorf("error unwrapping namespace: %s", err)
	}
	for k, v := range selectors {
		if strings.HasPrefix(k, "labels.") {
			k = strings.TrimPrefix(k, "labels.")
			namespace.Metadata.Labels[k] = v
		}
	}
	return nil
}

// UnwrapInto unwraps a NamespaceWrapper into a provided *Namespace.
// Other data types are not supported.
func (e *NamespaceWrapper) UnwrapInto(face interface{}) error {
	switch namespace := face.(type) {
	case *corev3.Namespace:
		return e.unwrapIntoNamespace(namespace)
	default:
		return fmt.Errorf("error unwrapping namespace: unsupported type: %T", namespace)
	}
}

// SQLParams serializes a NamespaceWrapper into a slice of parameters,
// suitable for passing to a postgresql query.
func (e *NamespaceWrapper) SQLParams() []interface{} {
	return []interface{}{
		&e.Name,
		&e.Selectors,
		&e.Annotations,
		&e.ID,
		&e.CreatedAt,
		&e.UpdatedAt,
		&e.DeletedAt,
	}
}

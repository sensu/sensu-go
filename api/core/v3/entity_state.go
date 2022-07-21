package v3

import (
	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

var entityStateRBACName = (&corev2.Entity{}).RBACName()

func (*EntityState) rbacName() string {
	return entityStateRBACName
}

func (e *EntityState) Fields() map[string]string {
	fields := map[string]string{
		"entity_state.name":      e.Metadata.Name,
		"entity_state.namespace": e.Metadata.Namespace,
	}
	return fields
}

package v3

import (
	"strconv"
	"strings"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

var entityConfigRBACName = (&corev2.Entity{}).RBACName()

func (e *EntityConfig) rbacName() string {
	return entityConfigRBACName
}

func (e *EntityConfig) Fields() map[string]string {
	fields := map[string]string{
		"entity_config.name":          e.Metadata.Name,
		"entity_config.namespace":     e.Metadata.Namespace,
		"entity_config.deregister":    strconv.FormatBool(e.Deregister),
		"entity_config.entity_class":  e.EntityClass,
		"entity_config.subscriptions": strings.Join(e.Subscriptions, ","),
	}
	MergeMapWithPrefix(fields, e.Metadata.Labels, "entity_config.labels.")
	return fields
}

// MergeMapWithPrefix merges contents of one map into another using a prefix.
func MergeMapWithPrefix(a map[string]string, b map[string]string, prefix string) {
	for k, v := range b {
		a[prefix+k] = v
	}
}

package v3

import (
	"strconv"
	"strings"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	stringutil "github.com/sensu/sensu-go/api/core/v3/internal/strings"
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

func redactMap(m map[string]string, redact []string) map[string]string {
	if len(redact) == 0 {
		redact = corev2.DefaultRedactFields
	}
	result := make(map[string]string, len(m))
	for k, v := range m {
		if stringutil.FoundInArray(k, redact) {
			result[k] = corev2.Redacted
		} else {
			result[k] = v
		}
	}
	return result
}

// ProduceRedacted redacts the entity according to the entity's Redact fields.
// A redacted copy is returned. The copy contains pointers to the original's
// memory, with different Labels and Annotations.
func (e *EntityConfig) ProduceRedacted() Resource {
	if e == nil {
		return nil
	}
	if e.Metadata == nil || (e.Metadata.Labels == nil && e.Metadata.Annotations == nil) {
		return e
	}
	copy := &EntityConfig{}
	*copy = *e
	copy.Metadata.Annotations = redactMap(e.Metadata.Annotations, e.Redact)
	copy.Metadata.Labels = redactMap(e.Metadata.Labels, e.Redact)
	return copy
}

package util_api

import (
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
)

// InputToTypeMeta translates TypeMetaInput -> corev2.TypeMeta
func InputToTypeMeta(i *schema.TypeMetaInput) *corev2.TypeMeta {
	if i == nil {
		return nil
	}
	return &corev2.TypeMeta{
		Type:       i.Type,
		APIVersion: i.ApiVersion,
	}
}

// ToObjectMeta returns ObjectMeta for given
func ToObjectMeta(m interface{}) corev2.ObjectMeta {
	if w, ok := m.(interface{ GetObjectMeta() corev2.ObjectMeta }); ok {
		m = w.GetObjectMeta()
	} else if w, ok := m.(interface{ GetMetadata() *corev2.ObjectMeta }); ok {
		m = w.GetMetadata()
	}
	switch m := m.(type) {
	case corev2.ObjectMeta:
		return m
	case *corev2.ObjectMeta:
		return *m
	}
	return corev2.ObjectMeta{}
}

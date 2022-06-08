package v3

import (
	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// Namespace resource contains standard resource object metadata for
// a given namespace.
type Namespace struct {
	Metadata *corev2.ObjectMeta `json:"metadata"`
}

func (n *Namespace) IsGlobalResource() bool {
	return true
}

func (n *Namespace) GetMetadata() *corev2.ObjectMeta {
	return n.Metadata
}

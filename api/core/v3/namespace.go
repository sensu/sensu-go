package v3

import (
	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

const (
	NamespaceSortName = "NAME"
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

func FixtureNamespace(name string) *Namespace {
	return &Namespace{
		Metadata: &corev2.ObjectMeta{
			Namespace:   "",
			Name:        name,
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
		},
	}
}

func V2NamespaceToV3(n *corev2.Namespace) *Namespace {
	return &Namespace{
		Metadata: &corev2.ObjectMeta{
			Namespace:   "",
			Name:        n.Name,
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
		},
	}
}

func V3NamespaceToV2(n *Namespace) *corev2.Namespace {
	return &corev2.Namespace{
		Name: n.Metadata.Name,
	}
}

func NewNamespace(name string) *Namespace {
	return &Namespace{
		Metadata: &corev2.ObjectMeta{
			Namespace:   "",
			Name:        name,
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
		},
	}
}

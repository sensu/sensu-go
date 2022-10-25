package postgres

import (
	"context"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/store"
)

// NamespaceStoreV1 is a storev1 compatibility shim which uses a storev2
// NamespaceStore as a backend
type NamespaceStoreV1 struct {
	namespaceStore *NamespaceStore
}

func NewNamespaceStoreV1(namespaceStore *NamespaceStore) *NamespaceStoreV1 {
	return &NamespaceStoreV1{
		namespaceStore: namespaceStore,
	}
}

func (s *NamespaceStoreV1) CreateNamespace(ctx context.Context, namespace *corev2.Namespace) error {
	return s.namespaceStore.CreateIfNotExists(ctx, corev3.V2NamespaceToV3(namespace))
}

func (s *NamespaceStoreV1) DeleteNamespace(ctx context.Context, name string) error {
	return s.namespaceStore.Delete(ctx, name)
}

func (s *NamespaceStoreV1) GetNamespace(ctx context.Context, name string) (*corev2.Namespace, error) {
	namespace, err := s.namespaceStore.Get(ctx, name)
	if err != nil {
		return nil, err
	}
	return corev3.V3NamespaceToV2(namespace), nil
}

func (s *NamespaceStoreV1) ListNamespaces(ctx context.Context, pred *store.SelectionPredicate) ([]*corev2.Namespace, error) {
	namespacesV2 := []*corev2.Namespace{}
	namespaces, err := s.namespaceStore.List(ctx, pred)
	if err != nil {
		return nil, err
	}
	for _, namespace := range namespaces {
		namespacesV2 = append(namespacesV2, corev3.V3NamespaceToV2(namespace))
	}
	return namespacesV2, nil
}

func (s *NamespaceStoreV1) UpdateNamespace(ctx context.Context, namespace *corev2.Namespace) error {
	return s.namespaceStore.CreateOrUpdate(ctx, corev3.V2NamespaceToV3(namespace))
}

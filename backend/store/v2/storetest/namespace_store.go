package storetest

import (
	"context"

	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/patch"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/stretchr/testify/mock"
)

var _ storev2.NamespaceStore = new(NamespaceStore)

type NamespaceStore struct {
	mock.Mock
}

func (s *NamespaceStore) CreateIfNotExists(ctx context.Context, namespace *corev3.Namespace) error {
	args := s.Called(ctx, namespace)
	return args.Error(0)
}

func (s *NamespaceStore) CreateOrUpdate(ctx context.Context, namespace *corev3.Namespace) error {
	args := s.Called(ctx, namespace)
	return args.Error(0)
}

func (s *NamespaceStore) Delete(ctx context.Context, namespace string) error {
	args := s.Called(ctx, namespace)
	return args.Error(0)
}

func (s *NamespaceStore) Exists(ctx context.Context, namespace string) (bool, error) {
	args := s.Called(ctx, namespace)
	return args.Get(0).(bool), args.Error(1)
}

func (s *NamespaceStore) Get(ctx context.Context, namespace string) (*corev3.Namespace, error) {
	args := s.Called(ctx, namespace)
	w, _ := args.Get(0).(*corev3.Namespace)
	return w, args.Error(1)
}

func (s *NamespaceStore) IsEmpty(ctx context.Context, namespace string) (bool, error) {
	args := s.Called(ctx, namespace)
	return args.Get(0).(bool), args.Error(1)
}

func (s *NamespaceStore) List(ctx context.Context, pred *store.SelectionPredicate) ([]*corev3.Namespace, error) {
	args := s.Called(ctx, pred)
	return args.Get(0).([]*corev3.Namespace), args.Error(1)
}

func (s *NamespaceStore) Count(ctx context.Context) (int, error) {
	args := s.Called(ctx)
	return args.Get(0).(int), args.Error(1)
}
func (s *NamespaceStore) Patch(ctx context.Context, namespace string, patcher patch.Patcher, conditions *store.ETagCondition) error {
	args := s.Called(ctx, namespace, patcher, conditions)
	return args.Error(0)
}

func (s *NamespaceStore) UpdateIfExists(ctx context.Context, namespace *corev3.Namespace) error {
	args := s.Called(ctx, namespace)
	return args.Error(0)
}

package storetest

import (
	"context"

	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/patch"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/stretchr/testify/mock"
)

var _ storev2.EntityConfigStore = new(EntityConfigStore)

type EntityConfigStore struct {
	mock.Mock
}

func (s *EntityConfigStore) CreateIfNotExists(ctx context.Context, entityConfig *corev3.EntityConfig) error {
	args := s.Called(ctx, entityConfig)
	return args.Error(0)
}

func (s *EntityConfigStore) CreateOrUpdate(ctx context.Context, entityConfig *corev3.EntityConfig) error {
	args := s.Called(ctx, entityConfig)
	return args.Error(0)
}

func (s *EntityConfigStore) Delete(ctx context.Context, namespace, name string) error {
	args := s.Called(ctx, namespace, name)
	return args.Error(0)
}

func (s *EntityConfigStore) Exists(ctx context.Context, namespace, name string) (bool, error) {
	args := s.Called(ctx, namespace, name)
	return args.Get(0).(bool), args.Error(1)
}

func (s *EntityConfigStore) Get(ctx context.Context, namespace, name string) (*corev3.EntityConfig, error) {
	args := s.Called(ctx, namespace, name)
	w, _ := args.Get(0).(*corev3.EntityConfig)
	return w, args.Error(1)
}

func (s *EntityConfigStore) IsEmpty(ctx context.Context, namespace, name string) (bool, error) {
	args := s.Called(ctx, namespace, name)
	return args.Get(0).(bool), args.Error(1)
}

func (s *EntityConfigStore) List(ctx context.Context, namespace string, pred *store.SelectionPredicate) ([]*corev3.EntityConfig, error) {
	args := s.Called(ctx, namespace, pred)
	return args.Get(0).([]*corev3.EntityConfig), args.Error(1)
}

func (s *EntityConfigStore) Patch(ctx context.Context, namespace, entity string, patcher patch.Patcher, conditions *store.ETagCondition) error {
	args := s.Called(ctx, namespace, entity, patcher, conditions)
	return args.Error(0)
}

func (s *EntityConfigStore) UpdateIfExists(ctx context.Context, entityConfig *corev3.EntityConfig) error {
	args := s.Called(ctx, entityConfig)
	return args.Error(0)
}

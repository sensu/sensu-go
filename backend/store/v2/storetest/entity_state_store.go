package storetest

import (
	"context"

	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/patch"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/stretchr/testify/mock"
)

var _ storev2.EntityStateStore = new(EntityStateStore)

type EntityStateStore struct {
	mock.Mock
}

func (s *EntityStateStore) CreateIfNotExists(ctx context.Context, entityState *corev3.EntityState) error {
	args := s.Called(ctx, entityState)
	return args.Error(0)
}

func (s *EntityStateStore) CreateOrUpdate(ctx context.Context, entityState *corev3.EntityState) error {
	args := s.Called(ctx, entityState)
	return args.Error(0)
}

func (s *EntityStateStore) Delete(ctx context.Context, namespace, name string) error {
	args := s.Called(ctx, namespace, name)
	return args.Error(0)
}

func (s *EntityStateStore) Exists(ctx context.Context, namespace, name string) (bool, error) {
	args := s.Called(ctx, namespace, name)
	return args.Get(0).(bool), args.Error(1)
}

func (s *EntityStateStore) Get(ctx context.Context, namespace, name string) (*corev3.EntityState, error) {
	args := s.Called(ctx, namespace, name)
	w, _ := args.Get(0).(*corev3.EntityState)
	return w, args.Error(1)
}

func (s *EntityStateStore) IsEmpty(ctx context.Context, namespace, name string) (bool, error) {
	args := s.Called(ctx, namespace, name)
	return args.Get(0).(bool), args.Error(1)
}

func (s *EntityStateStore) List(ctx context.Context, namespace string, pred *store.SelectionPredicate) ([]*corev3.EntityState, error) {
	args := s.Called(ctx, namespace, pred)
	return args.Get(0).([]*corev3.EntityState), args.Error(1)
}

func (s *EntityStateStore) Patch(ctx context.Context, namespace, entity string, patcher patch.Patcher, conditions *store.ETagCondition) error {
	args := s.Called(ctx, namespace, entity, patcher, conditions)
	return args.Error(0)
}

func (s *EntityStateStore) UpdateIfExists(ctx context.Context, entityState *corev3.EntityState) error {
	args := s.Called(ctx, entityState)
	return args.Error(0)
}

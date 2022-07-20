package storetest

import (
	"context"

	"github.com/stretchr/testify/mock"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/patch"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

var _ storev2.Interface = new(Store)
var _ storev2.KeepaliveStore = new(KeepaliveStore)

type Store struct {
	mock.Mock
}

func (s *Store) CreateOrUpdate(ctx context.Context, req storev2.ResourceRequest, w storev2.Wrapper) error {
	args := s.Called(ctx, req, w)
	return args.Error(0)
}

func (s *Store) UpdateIfExists(ctx context.Context, req storev2.ResourceRequest, w storev2.Wrapper) error {
	args := s.Called(ctx, req, w)
	return args.Error(0)
}

func (s *Store) CreateIfNotExists(ctx context.Context, req storev2.ResourceRequest, w storev2.Wrapper) error {
	args := s.Called(ctx, req, w)
	return args.Error(0)
}

func (s *Store) Get(ctx context.Context, req storev2.ResourceRequest) (storev2.Wrapper, error) {
	args := s.Called(ctx, req)
	w, _ := args.Get(0).(storev2.Wrapper)
	return w, args.Error(1)
}

func (s *Store) Delete(ctx context.Context, req storev2.ResourceRequest) error {
	args := s.Called(ctx, req)
	return args.Error(0)
}

func (s *Store) List(ctx context.Context, req storev2.ResourceRequest, pred *store.SelectionPredicate) (storev2.WrapList, error) {
	args := s.Called(ctx, req, pred)
	return args.Get(0).(storev2.WrapList), args.Error(1)
}

func (s *Store) Exists(ctx context.Context, req storev2.ResourceRequest) (bool, error) {
	args := s.Called(ctx, req)
	return args.Get(0).(bool), args.Error(1)
}

func (s *Store) Patch(ctx context.Context, req storev2.ResourceRequest, w storev2.Wrapper, patcher patch.Patcher, conditions *store.ETagCondition) error {
	args := s.Called(ctx, req, w, patcher, conditions)
	return args.Error(0)
}

func (s *Store) Watch(ctx context.Context, req storev2.ResourceRequest) <-chan []storev2.WatchEvent {
	return s.Called(ctx, req).Get(0).(<-chan []storev2.WatchEvent)
}

func (s *Store) NamespaceStore() storev2.NamespaceStore {
	return s.Called().Get(0).(storev2.NamespaceStore)
}

func (s *Store) EntityConfigStore() storev2.EntityConfigStore {
	return s.Called().Get(0).(storev2.EntityConfigStore)
}

func (s *Store) EntityStateStore() storev2.EntityStateStore {
	return s.Called().Get(0).(storev2.EntityStateStore)
}

func (s *Store) Initialize(ctx context.Context, fn storev2.InitializeFunc) error {
	args := s.Called(ctx, fn)
	return args.Error(0)
}

type KeepaliveStore struct {
	mock.Mock
}

func (s *KeepaliveStore) DeleteFailingKeepalive(ctx context.Context, entityConfig *corev3.EntityConfig) error {
	args := s.Called(ctx, entityConfig)
	return args.Error(0)
}

func (s *KeepaliveStore) GetFailingKeepalives(ctx context.Context) ([]*corev2.KeepaliveRecord, error) {
	args := s.Called(ctx)
	return args.Get(0).([]*corev2.KeepaliveRecord), args.Error(1)
}

func (s *KeepaliveStore) UpdateFailingKeepalive(ctx context.Context, entityConfig *corev3.EntityConfig, expiration int64) error {
	args := s.Called(ctx, entityConfig, expiration)
	return args.Error(0)
}

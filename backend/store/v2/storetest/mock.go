package storetest

import (
	"context"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/patch"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/stretchr/testify/mock"
)

var _ storev2.Interface = new(Store)

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

func (s *Store) Initialize(ctx context.Context, fn storev2.InitializeFunc) error {
	args := s.Called(ctx, fn)
	return args.Error(0)
}

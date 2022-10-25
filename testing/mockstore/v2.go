package mockstore

import (
	"context"
	"reflect"

	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/patch"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/stretchr/testify/mock"
)

type V2MockStore struct {
	mock.Mock
}

func (v *V2MockStore) CreateOrUpdate(ctx context.Context, req storev2.ResourceRequest, w storev2.Wrapper) error {
	return v.Called(ctx, req, w).Error(0)
}

func (v *V2MockStore) UpdateIfExists(ctx context.Context, req storev2.ResourceRequest, w storev2.Wrapper) error {
	return v.Called(ctx, req, w).Error(0)
}

func (v *V2MockStore) CreateIfNotExists(ctx context.Context, req storev2.ResourceRequest, w storev2.Wrapper) error {
	return v.Called(ctx, req, w).Error(0)
}

func (v *V2MockStore) Get(ctx context.Context, req storev2.ResourceRequest) (storev2.Wrapper, error) {
	args := v.Called(ctx, req)
	wrapper, _ := args.Get(0).(storev2.Wrapper)
	return wrapper, args.Error(1)
}

func (v *V2MockStore) Delete(ctx context.Context, req storev2.ResourceRequest) error {
	return v.Called(ctx, req).Error(0)
}

func (v *V2MockStore) List(ctx context.Context, req storev2.ResourceRequest, pred *store.SelectionPredicate) (storev2.WrapList, error) {
	args := v.Called(ctx, req, pred)
	list, _ := args.Get(0).(storev2.WrapList)
	return list, args.Error(1)
}

func (v *V2MockStore) Exists(ctx context.Context, req storev2.ResourceRequest) (bool, error) {
	args := v.Called(ctx, req)
	return args.Get(0).(bool), args.Error(1)
}

func (v *V2MockStore) Patch(ctx context.Context, req storev2.ResourceRequest, w storev2.Wrapper, patcher patch.Patcher, cond *store.ETagCondition) error {
	return v.Called(ctx, req, w, patcher, cond).Error(0)
}

func (v *V2MockStore) Watch(ctx context.Context, req storev2.ResourceRequest) <-chan []storev2.WatchEvent {
	args := v.Called(ctx, req)
	return args.Get(0).(<-chan []storev2.WatchEvent)
}

func (v *V2MockStore) NamespaceStore() storev2.NamespaceStore {
	args := v.Called()
	return args.Get(0).(storev2.NamespaceStore)
}

func (v *V2MockStore) EntityConfigStore() storev2.EntityConfigStore {
	args := v.Called()
	return args.Get(0).(storev2.EntityConfigStore)
}

func (v *V2MockStore) EntityStateStore() storev2.EntityStateStore {
	args := v.Called()
	return args.Get(0).(storev2.EntityStateStore)
}

func (v *V2MockStore) Initialize(ctx context.Context, fn storev2.InitializeFunc) error {
	args := v.Called(ctx, fn)
	return args.Error(1)
}

type WrapList[T corev3.Resource] []T

func (w WrapList[T]) Unwrap() ([]corev3.Resource, error) {
	result := make([]corev3.Resource, 0)
	for _, resource := range w {
		result = append(result, resource)
	}
	return result, nil
}

func (w WrapList[T]) UnwrapInto(target interface{}) error {
	list, ok := target.(*[]T)
	if !ok {
		panic("bad target")
	}
	*list = w
	return nil
}

func (w WrapList[T]) Len() int {
	return len(w)
}

type Wrapper[T corev3.Resource] struct {
	Value T
}

func (w Wrapper[T]) Unwrap() (corev3.Resource, error) {
	return w.Value, nil
}

func (w Wrapper[T]) UnwrapInto(target interface{}) error {
	val, ok := target.(T)
	if !ok {
		panic("bad target")
	}
	if !reflect.ValueOf(w.Value).IsZero() {
		reflect.ValueOf(val).Elem().Set(reflect.ValueOf(w.Value).Elem())
	}
	return nil
}

package mockstore

import (
	"context"
	"errors"
	"reflect"

	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/patch"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/stretchr/testify/mock"
)

type V2MockStore struct {
	mock.Mock
}

func (v *V2MockStore) CreateOrUpdate(req storev2.ResourceRequest, w storev2.Wrapper) error {
	return v.Called(req, w).Error(0)
}

func (v *V2MockStore) UpdateIfExists(req storev2.ResourceRequest, w storev2.Wrapper) error {
	return v.Called(req, w).Error(0)
}

func (v *V2MockStore) CreateIfNotExists(req storev2.ResourceRequest, w storev2.Wrapper) error {
	return v.Called(req, w).Error(0)
}

func (v *V2MockStore) Get(req storev2.ResourceRequest) (storev2.Wrapper, error) {
	args := v.Called(req)
	wrapper, _ := args.Get(0).(storev2.Wrapper)
	return wrapper, args.Error(1)
}

func (v *V2MockStore) Delete(req storev2.ResourceRequest) error {
	return v.Called(req).Error(0)
}

func (v *V2MockStore) List(req storev2.ResourceRequest, pred *store.SelectionPredicate) (storev2.WrapList, error) {
	args := v.Called(req, pred)
	list, _ := args.Get(0).(storev2.WrapList)
	return list, args.Error(1)
}

func (v *V2MockStore) Exists(req storev2.ResourceRequest) (bool, error) {
	args := v.Called(req)
	return args.Get(0).(bool), args.Error(1)
}

func (v *V2MockStore) Patch(req storev2.ResourceRequest, w storev2.Wrapper, patcher patch.Patcher, cond *store.ETagCondition) error {
	return v.Called(req, w, patcher, cond).Error(0)
}

func (v *V2MockStore) Watch(ctx context.Context, req storev2.ResourceRequest) <-chan []storev2.WatchEvent {
	args := v.Called(ctx, req)
	return args.Get(0).(<-chan []storev2.WatchEvent)
}

func (v *V2MockStore) CreateNamespace(ctx context.Context, ns *corev3.Namespace) error {
	return v.Called(ctx, ns).Error(0)
}

func (v *V2MockStore) DeleteNamespace(ctx context.Context, name string) error {
	return v.Called(ctx, name).Error(0)
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
		return errors.New("bad target")
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
	reflect.ValueOf(val).Elem().Set(reflect.ValueOf(w.Value).Elem())
	return nil
}

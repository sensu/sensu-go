package mockstore

import (
	"context"
	"fmt"
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

func (v *V2MockStore) GetEntityConfigStore() storev2.EntityConfigStore {
	return v.Called().Get(0).(storev2.EntityConfigStore)
}

func (v *V2MockStore) GetEntityStateStore() storev2.EntityStateStore {
	return v.Called().Get(0).(storev2.EntityStateStore)
}

func (v *V2MockStore) GetNamespaceStore() storev2.NamespaceStore {
	return v.Called().Get(0).(storev2.NamespaceStore)
}

func (v *V2MockStore) GetConfigStore() storev2.ConfigStore {
	return v.Called().Get(0).(storev2.ConfigStore)
}

func (v *V2MockStore) GetEventStore() store.EventStore {
	return v.Called().Get(0).(store.EventStore)
}

func (v *V2MockStore) GetEntityStore() store.EntityStore {
	return v.Called().Get(0).(store.EntityStore)
}

type ConfigStore struct {
	mock.Mock
}

func (v *ConfigStore) CreateOrUpdate(ctx context.Context, req storev2.ResourceRequest, w storev2.Wrapper) error {
	return v.Called(ctx, req, w).Error(0)
}

func (v *ConfigStore) UpdateIfExists(ctx context.Context, req storev2.ResourceRequest, w storev2.Wrapper) error {
	return v.Called(ctx, req, w).Error(0)
}

func (v *ConfigStore) CreateIfNotExists(ctx context.Context, req storev2.ResourceRequest, w storev2.Wrapper) error {
	return v.Called(ctx, req, w).Error(0)
}

func (v *ConfigStore) Get(ctx context.Context, req storev2.ResourceRequest) (storev2.Wrapper, error) {
	args := v.Called(ctx, req)
	wrapper, _ := args.Get(0).(storev2.Wrapper)
	return wrapper, args.Error(1)
}

func (v *ConfigStore) Delete(ctx context.Context, req storev2.ResourceRequest) error {
	return v.Called(ctx, req).Error(0)
}

func (v *ConfigStore) List(ctx context.Context, req storev2.ResourceRequest, pred *store.SelectionPredicate) (storev2.WrapList, error) {
	args := v.Called(ctx, req, pred)
	list, _ := args.Get(0).(storev2.WrapList)
	return list, args.Error(1)
}

func (v *ConfigStore) Count(ctx context.Context, req storev2.ResourceRequest) (int, error) {
	args := v.Called(ctx, req)
	return args.Get(0).(int), args.Error(1)
}

func (v *ConfigStore) Exists(ctx context.Context, req storev2.ResourceRequest) (bool, error) {
	args := v.Called(ctx, req)
	return args.Get(0).(bool), args.Error(1)
}

func (v *ConfigStore) Patch(ctx context.Context, req storev2.ResourceRequest, w storev2.Wrapper, patcher patch.Patcher, cond *store.ETagCondition) error {
	return v.Called(ctx, req, w, patcher, cond).Error(0)
}

func (v *ConfigStore) Watch(ctx context.Context, req storev2.ResourceRequest) <-chan []storev2.WatchEvent {
	args := v.Called(ctx, req)
	return args.Get(0).(<-chan []storev2.WatchEvent)
}

func (v *ConfigStore) NamespaceStore() storev2.NamespaceStore {
	args := v.Called()
	return args.Get(0).(storev2.NamespaceStore)
}

func (v *ConfigStore) EntityConfigStore() storev2.EntityConfigStore {
	args := v.Called()
	return args.Get(0).(storev2.EntityConfigStore)
}

func (v *ConfigStore) EntityStateStore() storev2.EntityStateStore {
	args := v.Called()
	return args.Get(0).(storev2.EntityStateStore)
}

func (v *ConfigStore) Initialize(ctx context.Context, fn storev2.InitializeFunc) error {
	args := v.Called(ctx, fn)
	return args.Error(0)
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
		target := fmt.Sprintf("%T", target)
		panic("bad target: " + target + ", want: " + fmt.Sprintf("%T", *new(T)))
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
		target := fmt.Sprintf("%T", target)
		panic("bad target: " + target + ", want: " + fmt.Sprintf("%T", *new(T)))
	}
	if !reflect.ValueOf(w.Value).IsZero() {
		reflect.ValueOf(val).Elem().Set(reflect.ValueOf(w.Value).Elem())
	}
	return nil
}

type EntityStateStore struct {
	mock.Mock
}

func (s *EntityStateStore) Count(ctx context.Context, namespace string) (int, error) {
	args := s.Called(ctx, namespace)
	return args.Get(0).(int), args.Error(1)
}

func (e *EntityStateStore) CreateOrUpdate(ctx context.Context, ec *corev3.EntityState) error {
	return e.Called(ctx, ec).Error(0)
}

func (e *EntityStateStore) UpdateIfExists(ctx context.Context, ec *corev3.EntityState) error {
	return e.Called(ctx, ec).Error(0)
}

func (e *EntityStateStore) CreateIfNotExists(ctx context.Context, ec *corev3.EntityState) error {
	return e.Called(ctx, ec).Error(0)
}

func (e *EntityStateStore) Get(ctx context.Context, ns string, n string) (*corev3.EntityState, error) {
	args := e.Called(ctx, ns, n)
	return args.Get(0).(*corev3.EntityState), args.Error(1)
}

func (e *EntityStateStore) Delete(ctx context.Context, ns string, n string) error {
	return e.Called(ctx, ns, n).Error(0)
}

func (e *EntityStateStore) List(ctx context.Context, ns string, pred *store.SelectionPredicate) ([]*corev3.EntityState, error) {
	args := e.Called(ctx, ns, pred)
	return args.Get(0).([]*corev3.EntityState), args.Error(1)
}

func (e *EntityStateStore) Exists(ctx context.Context, ns string, name string) (bool, error) {
	args := e.Called(ctx, ns, name)
	return args.Bool(0), args.Error(1)
}

func (e *EntityStateStore) Patch(ctx context.Context, ns string, n string, p patch.Patcher, etag *store.ETagCondition) error {
	return e.Called(ctx, ns, n, p, etag).Error(0)
}

type EntityConfigStore struct {
	mock.Mock
}

func (s *EntityConfigStore) Count(ctx context.Context, namespace string, entityClass string) (int, error) {
	args := s.Called(ctx, namespace, entityClass)
	return args.Get(0).(int), args.Error(1)
}

func (e *EntityConfigStore) CreateOrUpdate(ctx context.Context, ec *corev3.EntityConfig) error {
	return e.Called(ctx, ec).Error(0)
}

func (e *EntityConfigStore) UpdateIfExists(ctx context.Context, ec *corev3.EntityConfig) error {
	return e.Called(ctx, ec).Error(0)
}

func (e *EntityConfigStore) CreateIfNotExists(ctx context.Context, ec *corev3.EntityConfig) error {
	return e.Called(ctx, ec).Error(0)
}

func (e *EntityConfigStore) Get(ctx context.Context, ns string, n string) (*corev3.EntityConfig, error) {
	args := e.Called(ctx, ns, n)
	return args.Get(0).(*corev3.EntityConfig), args.Error(1)
}

func (e *EntityConfigStore) Delete(ctx context.Context, ns string, n string) error {
	return e.Called(ctx, ns, n).Error(0)
}

func (e *EntityConfigStore) List(ctx context.Context, ns string, pred *store.SelectionPredicate) ([]*corev3.EntityConfig, error) {
	args := e.Called(ctx, ns, pred)
	return args.Get(0).([]*corev3.EntityConfig), args.Error(1)
}

func (e *EntityConfigStore) Exists(ctx context.Context, ns string, name string) (bool, error) {
	args := e.Called(ctx, ns, name)
	return args.Bool(0), args.Error(1)
}

func (e *EntityConfigStore) Patch(ctx context.Context, ns string, n string, p patch.Patcher, etag *store.ETagCondition) error {
	return e.Called(ctx, ns, n, p, etag).Error(0)
}

func (e *EntityConfigStore) Watch(ctx context.Context, ns, n string) <-chan []storev2.WatchEvent {
	return e.Called(ctx, ns, n).Get(0).(<-chan []storev2.WatchEvent)
}

type NamespaceStore struct {
	mock.Mock
}

func (s *NamespaceStore) Count(ctx context.Context) (int, error) {
	args := s.Called(ctx)
	return args.Get(0).(int), args.Error(1)
}

func (n *NamespaceStore) CreateOrUpdate(ctx context.Context, ns *corev3.Namespace) error {
	return n.Called(ctx, ns).Error(0)
}

func (n *NamespaceStore) UpdateIfExists(ctx context.Context, ns *corev3.Namespace) error {
	return n.Called(ctx, ns).Error(0)
}

func (n *NamespaceStore) CreateIfNotExists(ctx context.Context, ns *corev3.Namespace) error {
	return n.Called(ctx, ns).Error(0)
}

func (n *NamespaceStore) Get(ctx context.Context, ns string) (*corev3.Namespace, error) {
	args := n.Called(ctx, ns)
	return args.Get(0).(*corev3.Namespace), args.Error(1)
}

func (n *NamespaceStore) Delete(ctx context.Context, ns string) error {
	return n.Called(ctx, ns).Error(0)
}

func (n *NamespaceStore) List(ctx context.Context, pred *store.SelectionPredicate) ([]*corev3.Namespace, error) {
	args := n.Called(ctx, pred)
	return args.Get(0).([]*corev3.Namespace), args.Error(1)
}

func (n *NamespaceStore) Exists(ctx context.Context, ns string) (bool, error) {
	args := n.Called(ctx, ns)
	return args.Bool(0), args.Error(1)
}

func (n *NamespaceStore) Patch(ctx context.Context, ns string, p patch.Patcher, pred *store.ETagCondition) error {
	return n.Called(ctx, ns, p, pred).Error(0)
}

func (n *NamespaceStore) IsEmpty(ctx context.Context, ns string) (bool, error) {
	args := n.Called(ctx, ns)
	return args.Bool(0), args.Error(1)
}

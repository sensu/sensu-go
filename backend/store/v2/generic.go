package v2

import (
	"context"
	"reflect"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/patch"
)

type ID = corev2.ObjectMeta

type Resource[T any] interface {
	corev3.Resource
	*T
}

// Generic is a generic store that sits on top of an Interface and provides
// a type safe set of methods for dealing with resource CRUD operations.
// It allows users to avoid dealing with the Wrapper abstraction, and handles
// list allocations optimally.
type Generic[R Resource[T], T any] struct {
	Interface Interface
}

func NewGenericStore[R Resource[T], T any](face Interface) Generic[R, T] {
	return Generic[R, T]{Interface: face}
}

func makeResource[R Resource[T], T any]() *T {
	var t T
	return &t
}

func prepare(resource corev3.Resource) (ResourceRequest, Wrapper, error) {
	req := NewResourceRequestFromResource(resource)
	wrapper, err := WrapResource(resource)
	return req, wrapper, err
}

func (g Generic[R, T]) CreateOrUpdate(ctx context.Context, resource R) error {
	req, wrapper, err := prepare(resource)
	if err != nil {
		return err
	}
	return g.Interface.CreateOrUpdate(ctx, req, wrapper)
}

func (g Generic[R, T]) UpdateIfExists(ctx context.Context, resource R) error {
	req, wrapper, err := prepare(resource)
	if err != nil {
		return err
	}
	return g.Interface.UpdateIfExists(ctx, req, wrapper)
}

func (g Generic[R, T]) CreateIfNotExists(ctx context.Context, resource R) error {
	req, wrapper, err := prepare(resource)
	if err != nil {
		return err
	}
	return g.Interface.CreateIfNotExists(ctx, req, wrapper)
}

func (g Generic[R, T]) Get(ctx context.Context, id ID) (R, error) {
	result := makeResource[R, T]()
	tm := getGenericTypeMeta[R, T]()
	var r R
	req := NewResourceRequest(tm, id.Namespace, id.Name, r.StoreName())
	wrapper, err := g.Interface.Get(ctx, req)
	if err != nil {
		return nil, err
	}
	if err := wrapper.UnwrapInto(result); err != nil {
		return nil, err
	}
	return result, nil
}

func (g Generic[R, T]) Delete(ctx context.Context, id ID) error {
	var r R
	tm := getGenericTypeMeta[R, T]()
	req := NewResourceRequest(tm, id.Namespace, id.Name, r.StoreName())
	req.Namespace = id.Namespace
	req.Name = id.Name
	return g.Interface.Delete(ctx, req)
}

func (g Generic[R, T]) List(ctx context.Context, id ID, pred *store.SelectionPredicate) ([]T, error) {
	var r R
	tm := getGenericTypeMeta[R, T]()
	req := NewResourceRequest(tm, id.Namespace, "", r.StoreName())
	wrapper, err := g.Interface.List(ctx, req, pred)
	if err != nil {
		return nil, err
	}
	wlen := wrapper.Len()
	result := make([]T, wlen, wlen)
	if err := wrapper.UnwrapInto(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (g Generic[R, T]) Exists(ctx context.Context, id ID) (bool, error) {
	tm := getGenericTypeMeta[R, T]()
	var r R
	req := NewResourceRequest(tm, id.Namespace, id.Name, r.StoreName())
	return g.Interface.Exists(ctx, req)
}

func (g Generic[R, T]) Patch(ctx context.Context, resource R, patcher patch.Patcher, etag *store.ETagCondition) error {
	req, wrapper, err := prepare(resource)
	if err != nil {
		return err
	}
	return g.Interface.Patch(ctx, req, wrapper, patcher, etag)
}

func getGenericTypeMeta[R Resource[T], T any]() corev2.TypeMeta {
	var t T
	var tm corev2.TypeMeta
	if getter, ok := (interface{}(&t)).(tmGetter); ok {
		tm = getter.GetTypeMeta()
	} else {
		typ := reflect.TypeOf(t)
		tm = corev2.TypeMeta{
			Type:       typ.Name(),
			APIVersion: apiVersion(typ.PkgPath()),
		}
	}
	return tm
}

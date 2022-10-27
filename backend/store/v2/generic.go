package v2

import (
	"context"
	"errors"
	"reflect"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
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
//
// If Interface can be type asserted to be one of a EntityConfigStoreGetter,
// EntityStateStoreGetter, or NamespaceStoreGetter, then operations on those
// types will be dispatched to the store that the getter returns.
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

var errNoSpecialization = errors.New("no specialization available")

func (g Generic[R, T]) trySpecializeCreateOrUpdate(ctx context.Context, resource R) error {
	switch value := any(resource).(type) {
	case *corev3.EntityConfig:
		if getter, ok := g.Interface.(EntityConfigStoreGetter); ok {
			return getter.GetEntityConfigStore().CreateOrUpdate(ctx, value)
		}
		return errNoSpecialization
	case *corev3.EntityState:
		if getter, ok := g.Interface.(EntityStateStoreGetter); ok {
			return getter.GetEntityStateStore().CreateOrUpdate(ctx, value)
		}
		return errNoSpecialization
	case *corev3.Namespace:
		if getter, ok := g.Interface.(NamespaceStoreGetter); ok {
			return getter.GetNamespaceStore().CreateOrUpdate(ctx, value)
		}
		return errNoSpecialization
	default:
		return errNoSpecialization
	}
}

func (g Generic[R, T]) CreateOrUpdate(ctx context.Context, resource R) error {
	// try specialized path first
	if err := g.trySpecializeCreateOrUpdate(ctx, resource); err != nil {
		if err != errNoSpecialization {
			return err
		}
	} else {
		return nil
	}
	// fall back to common path
	req, wrapper, err := prepare(resource)
	if err != nil {
		return err
	}
	return g.Interface.CreateOrUpdate(ctx, req, wrapper)
}

func (g Generic[R, T]) trySpecializeUpdateIfExists(ctx context.Context, resource R) error {
	switch value := any(resource).(type) {
	case *corev3.EntityConfig:
		if getter, ok := g.Interface.(EntityConfigStoreGetter); ok {
			return getter.GetEntityConfigStore().UpdateIfExists(ctx, value)
		}
		return errNoSpecialization
	case *corev3.EntityState:
		if getter, ok := g.Interface.(EntityStateStoreGetter); ok {
			return getter.GetEntityStateStore().UpdateIfExists(ctx, value)
		}
		return errNoSpecialization
	case *corev3.Namespace:
		if getter, ok := g.Interface.(NamespaceStoreGetter); ok {
			return getter.GetNamespaceStore().UpdateIfExists(ctx, value)
		}
		return errNoSpecialization
	default:
		return errNoSpecialization
	}
}

func (g Generic[R, T]) UpdateIfExists(ctx context.Context, resource R) error {
	// try specialized path first
	if err := g.trySpecializeUpdateIfExists(ctx, resource); err != nil {
		if err != errNoSpecialization {
			return err
		}
	} else {
		return nil
	}
	// fall back to common path
	req, wrapper, err := prepare(resource)
	if err != nil {
		return err
	}
	return g.Interface.UpdateIfExists(ctx, req, wrapper)
}

func (g Generic[R, T]) trySpecializeCreateIfNotExists(ctx context.Context, resource R) error {
	switch value := any(resource).(type) {
	case *corev3.EntityConfig:
		if getter, ok := g.Interface.(EntityConfigStoreGetter); ok {
			return getter.GetEntityConfigStore().CreateIfNotExists(ctx, value)
		}
		return errNoSpecialization
	case *corev3.EntityState:
		if getter, ok := g.Interface.(EntityStateStoreGetter); ok {
			return getter.GetEntityStateStore().CreateIfNotExists(ctx, value)
		}
		return errNoSpecialization
	case *corev3.Namespace:
		if getter, ok := g.Interface.(NamespaceStoreGetter); ok {
			return getter.GetNamespaceStore().CreateIfNotExists(ctx, value)
		}
		return errNoSpecialization
	default:
		return errNoSpecialization
	}
}

func (g Generic[R, T]) CreateIfNotExists(ctx context.Context, resource R) error {
	// try specialized path first
	if err := g.trySpecializeCreateIfNotExists(ctx, resource); err != nil {
		if err != errNoSpecialization {
			return err
		}
	} else {
		return nil
	}
	// common path
	req, wrapper, err := prepare(resource)
	if err != nil {
		return err
	}
	return g.Interface.CreateIfNotExists(ctx, req, wrapper)
}

func (g Generic[R, T]) trySpecializeGet(ctx context.Context, id ID) (R, error) {
	switch any(new(T)).(type) {
	case *corev3.EntityConfig:
		if getter, ok := g.Interface.(EntityConfigStoreGetter); ok {
			val, err := getter.GetEntityConfigStore().Get(ctx, id.Namespace, id.Name)
			return any(val).(R), err
		}
		return *new(R), errNoSpecialization
	case *corev3.EntityState:
		if getter, ok := g.Interface.(EntityStateStoreGetter); ok {
			val, err := getter.GetEntityStateStore().Get(ctx, id.Namespace, id.Name)
			return any(val).(R), err
		}
		return *new(R), errNoSpecialization
	case *corev3.Namespace:
		if getter, ok := g.Interface.(NamespaceStoreGetter); ok {
			val, err := getter.GetNamespaceStore().Get(ctx, id.Name)
			return any(val).(R), err
		}
		return *new(R), errNoSpecialization
	default:
		return *new(R), errNoSpecialization
	}
}

func (g Generic[R, T]) Get(ctx context.Context, id ID) (R, error) {
	// try specialized path first
	if val, err := g.trySpecializeGet(ctx, id); err != nil {
		if err != errNoSpecialization {
			return val, err
		}
	} else {
		return val, nil
	}
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

func (g Generic[R, T]) trySpecializeDelete(ctx context.Context, id ID) error {
	switch any(new(T)).(type) {
	case *corev3.EntityConfig:
		if getter, ok := g.Interface.(EntityConfigStoreGetter); ok {
			return getter.GetEntityConfigStore().Delete(ctx, id.Namespace, id.Name)
		}
		return errNoSpecialization
	case *corev3.EntityState:
		if getter, ok := g.Interface.(EntityStateStoreGetter); ok {
			return getter.GetEntityStateStore().Delete(ctx, id.Namespace, id.Name)
		}
		return errNoSpecialization
	case *corev3.Namespace:
		if getter, ok := g.Interface.(NamespaceStoreGetter); ok {
			return getter.GetNamespaceStore().Delete(ctx, id.Name)
		}
		return errNoSpecialization
	default:
		return errNoSpecialization
	}
}

func (g Generic[R, T]) Delete(ctx context.Context, id ID) error {
	// try specialized path first
	if err := g.trySpecializeDelete(ctx, id); err != nil {
		if err != errNoSpecialization {
			return err
		}
	} else {
		return nil
	}
	// common path
	var r R
	tm := getGenericTypeMeta[R, T]()
	req := NewResourceRequest(tm, id.Namespace, id.Name, r.StoreName())
	req.Namespace = id.Namespace
	req.Name = id.Name
	return g.Interface.Delete(ctx, req)
}

func (g Generic[R, T]) trySpecializeList(ctx context.Context, id ID, pred *store.SelectionPredicate) ([]T, error) {
	switch any(*new(R)).(type) {
	case *corev3.EntityConfig:
		if getter, ok := g.Interface.(EntityConfigStoreGetter); ok {
			values, err := getter.GetEntityConfigStore().List(ctx, id.Namespace, pred)
			if err != nil {
				return nil, err
			}
			result := make([]T, len(values))
			for i := range values {
				result[i] = any(*values[i]).(T)
			}
			return result, nil
		}
		return nil, errNoSpecialization
	case *corev3.EntityState:
		if getter, ok := g.Interface.(EntityStateStoreGetter); ok {
			values, err := getter.GetEntityStateStore().List(ctx, id.Namespace, pred)
			if err != nil {
				return nil, err
			}
			result := make([]T, len(values))
			for i := range values {
				result[i] = any(*values[i]).(T)
			}
			return result, nil
		}
		return nil, errNoSpecialization
	case *corev3.Namespace:
		if getter, ok := g.Interface.(NamespaceStoreGetter); ok {
			values, err := getter.GetNamespaceStore().List(ctx, pred)
			if err != nil {
				return nil, err
			}
			result := make([]T, len(values))
			for i := range values {
				result[i] = any(*values[i]).(T)
			}
			return result, nil
		}
		return nil, errNoSpecialization
	default:
		return nil, errNoSpecialization
	}

}

func (g Generic[R, T]) List(ctx context.Context, id ID, pred *store.SelectionPredicate) ([]T, error) {
	if lst, err := g.trySpecializeList(ctx, id, pred); err != nil {
		if err != errNoSpecialization {
			return nil, err
		}
	} else {
		return lst, nil
	}
	var r R
	tm := getGenericTypeMeta[R, T]()
	req := NewResourceRequest(tm, id.Namespace, "", r.StoreName())
	wrapper, err := g.Interface.List(ctx, req, pred)
	if err != nil {
		return nil, err
	}
	wlen := wrapper.Len()
	result := make([]T, wlen)
	if err := wrapper.UnwrapInto(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (g Generic[R, T]) trySpecializeExists(ctx context.Context, id ID) (bool, error) {
	switch any(new(T)).(type) {
	case *corev3.EntityConfig:
		if getter, ok := g.Interface.(EntityConfigStoreGetter); ok {
			return getter.GetEntityConfigStore().Exists(ctx, id.Namespace, id.Name)
		}
		return false, errNoSpecialization
	case *corev3.EntityState:
		if getter, ok := g.Interface.(EntityStateStoreGetter); ok {
			return getter.GetEntityStateStore().Exists(ctx, id.Namespace, id.Name)
		}
		return false, errNoSpecialization
	case *corev3.Namespace:
		if getter, ok := g.Interface.(NamespaceStoreGetter); ok {
			return getter.GetNamespaceStore().Exists(ctx, id.Name)
		}
		return false, errNoSpecialization
	default:
		return false, errNoSpecialization
	}
}

func (g Generic[R, T]) Exists(ctx context.Context, id ID) (bool, error) {
	if ok, err := g.trySpecializeExists(ctx, id); err != nil {
		if err != errNoSpecialization {
			return false, err
		}
	} else {
		return ok, nil
	}
	tm := getGenericTypeMeta[R, T]()
	var r R
	req := NewResourceRequest(tm, id.Namespace, id.Name, r.StoreName())
	return g.Interface.Exists(ctx, req)
}

func (g Generic[R, T]) trySpecializePatch(ctx context.Context, resource R, patcher patch.Patcher, etag *store.ETagCondition) error {
	meta := resource.GetMetadata()
	namespace := meta.Namespace
	name := meta.Name
	switch any(new(T)).(type) {
	case *corev3.EntityConfig:
		if getter, ok := g.Interface.(EntityConfigStoreGetter); ok {
			return getter.GetEntityConfigStore().Patch(ctx, namespace, name, patcher, etag)
		}
		return errNoSpecialization
	case *corev3.EntityState:
		if getter, ok := g.Interface.(EntityStateStoreGetter); ok {
			return getter.GetEntityStateStore().Patch(ctx, namespace, name, patcher, etag)
		}
		return errNoSpecialization
	case *corev3.Namespace:
		if getter, ok := g.Interface.(NamespaceStoreGetter); ok {
			return getter.GetNamespaceStore().Patch(ctx, name, patcher, etag)
		}
		return errNoSpecialization
	default:
		return errNoSpecialization
	}
}

func (g Generic[R, T]) Patch(ctx context.Context, resource R, patcher patch.Patcher, etag *store.ETagCondition) error {
	if err := g.trySpecializePatch(ctx, resource, patcher, etag); err != nil {
		if err != errNoSpecialization {
			return err
		}
	} else {
		return nil
	}
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

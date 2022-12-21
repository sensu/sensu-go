package v2

import (
	"context"
	"errors"
	"fmt"
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

func Of[R Resource[T], T any](face Interface) Generic[R, T] {
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

var (
	errNoSpecialization   = errors.New("no specialization available")
	errUnsupportedForType = errors.New("method not supported for type")
)

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
	return g.Interface.GetConfigStore().CreateOrUpdate(ctx, req, wrapper)
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
	return g.Interface.GetConfigStore().UpdateIfExists(ctx, req, wrapper)
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
	return g.Interface.GetConfigStore().CreateIfNotExists(ctx, req, wrapper)
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
	wrapper, err := g.Interface.GetConfigStore().Get(ctx, req)
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
	return g.Interface.GetConfigStore().Delete(ctx, req)
}

func (g Generic[R, T]) trySpecializeList(ctx context.Context, id ID, pred *store.SelectionPredicate) ([]R, error) {
	switch any(*new(R)).(type) {
	case *corev3.EntityConfig:
		if getter, ok := g.Interface.(EntityConfigStoreGetter); ok {
			values, err := getter.GetEntityConfigStore().List(ctx, id.Namespace, pred)
			if err != nil {
				return nil, err
			}
			result := make([]R, len(values))
			for i := range values {
				result[i] = any(values[i]).(R)
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
			result := make([]R, len(values))
			for i := range values {
				result[i] = any(values[i]).(R)
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
			result := make([]R, len(values))
			for i := range values {
				result[i] = any(values[i]).(R)
			}
			return result, nil
		}
		return nil, errNoSpecialization
	default:
		return nil, errNoSpecialization
	}

}

func (g Generic[R, T]) List(ctx context.Context, id ID, pred *store.SelectionPredicate) ([]R, error) {
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
	wrapper, err := g.Interface.GetConfigStore().List(ctx, req, pred)
	if err != nil {
		return nil, err
	}
	wlen := wrapper.Len()
	result := make([]R, wlen)
	if err := wrapper.UnwrapInto(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (g Generic[R, T]) trySpecializeCount(ctx context.Context, id ID) (int, error) {
	switch any(*new(R)).(type) {
	case *corev3.EntityConfig:
		if getter, ok := g.Interface.(EntityConfigStoreGetter); ok {
			return getter.GetEntityConfigStore().Count(ctx, id.Namespace, "")
		}
		return -1, errNoSpecialization
	case *corev3.EntityState:
		if getter, ok := g.Interface.(EntityStateStoreGetter); ok {
			return getter.GetEntityStateStore().Count(ctx, id.Namespace)
		}
		return -1, errNoSpecialization
	case *corev3.Namespace:
		if getter, ok := g.Interface.(NamespaceStoreGetter); ok {
			return getter.GetNamespaceStore().Count(ctx)
		}
		return -1, errNoSpecialization
	default:
		return -1, errNoSpecialization
	}
}

func (g Generic[R, T]) Count(ctx context.Context, id ID) (int, error) {
	if ct, err := g.trySpecializeCount(ctx, id); err != nil {
		if err != errNoSpecialization {
			return 0, err
		}
	} else {
		return ct, nil
	}
	var r R
	tm := getGenericTypeMeta[R, T]()
	req := NewResourceRequest(tm, id.Namespace, "", r.StoreName())
	return g.Interface.GetConfigStore().Count(ctx, req)
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
	return g.Interface.GetConfigStore().Exists(ctx, req)
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
	return g.Interface.GetConfigStore().Patch(ctx, req, wrapper, patcher, etag)
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

func (g Generic[R, T]) trySpecializeWatch(ctx context.Context, id ID) (<-chan []WatchEvent, error) {
	switch any(new(T)).(type) {
	case *corev3.EntityConfig:
		watch := g.Interface.GetEntityConfigStore().Watch(ctx, id.Name, id.Namespace)
		return watch, nil
	case *corev3.EntityState,
		*corev3.Namespace,
		*corev2.Entity,
		*corev2.Event,
		*corev2.Silenced:
		return nil, fmt.Errorf("%w: %T", errUnsupportedForType, new(T))
	}
	return nil, errNoSpecialization

}

func (g Generic[R, T]) Watch(ctx context.Context, id ID) <-chan []GenericEvent[R] {
	watch, err := g.trySpecializeWatch(ctx, id)
	if err != nil {
		if err != errNoSpecialization {
			errEvent := make(chan []GenericEvent[R], 1)
			errEvent <- []GenericEvent[R]{{Err: err}}
			close(errEvent)
			return errEvent
		}
	} else {
		return wrapWatch[R](ctx, watch)
	}

	var r R
	tm := getGenericTypeMeta[R]()
	req := NewResourceRequest(tm, id.Namespace, id.Name, r.StoreName())
	watch = g.Interface.GetConfigStore().Watch(ctx, req)
	return wrapWatch[R](ctx, watch)

}

type GenericEvent[R any] struct {
	Type          WatchActionType
	Key           corev2.ObjectMeta
	Value         R
	PreviousValue R
	Err           error
}

func wrapWatch[R Resource[T], T any](ctx context.Context, in <-chan []WatchEvent) <-chan []GenericEvent[R] {
	out := make(chan []GenericEvent[R], cap(in))
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case e, ok := <-in:
				if !ok {
					close(out)
					return
				}
				events := make([]GenericEvent[R], len(e))
				for i, event := range e {
					key := corev2.ObjectMeta{
						Name:      event.Key.Name,
						Namespace: event.Key.Namespace,
					}
					var resource, prev T
					if event.Value != nil {
						event.Value.UnwrapInto(&resource)
					}
					if event.PreviousValue != nil {
						event.PreviousValue.UnwrapInto(&prev)
					}

					events[i] = GenericEvent[R]{
						Type:          event.Type,
						Key:           key,
						Value:         &resource,
						PreviousValue: &prev,
						Err:           event.Err,
					}
				}
				out <- events
			}
		}
	}()

	return out
}

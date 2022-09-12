package postgres

import (
	"fmt"
	"sync"

	corev3 "github.com/sensu/sensu-go/api/core/v3"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/backend/store/v2/wrap"
)

type ResourceWrapper struct {
	mu           sync.RWMutex
	etcdWrapFunc func(corev3.Resource, ...wrap.Option) (storev2.Wrapper, error)
	pgEnabled    bool
}

func NewResourceWrapper(wrapFunc func(corev3.Resource, ...wrap.Option) (storev2.Wrapper, error)) *ResourceWrapper {
	return &ResourceWrapper{
		etcdWrapFunc: wrapFunc,
	}
}

func (e *ResourceWrapper) WrapResource(resource corev3.Resource, opts ...wrap.Option) (storev2.Wrapper, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.pgEnabled {
		return e.wrapWithPostgres(resource, opts...)
	}
	return e.etcdWrapFunc(resource, opts...)
}

func (e *ResourceWrapper) EnablePostgres() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.pgEnabled = true
}

func (e *ResourceWrapper) DisablePostgres() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.pgEnabled = false
}

func (e *ResourceWrapper) wrapWithPostgres(resource corev3.Resource, opts ...wrap.Option) (storev2.Wrapper, error) {
	switch value := resource.(type) {
	case *corev3.EntityConfig:
		return WrapEntityConfig(value), nil
	case *corev3.EntityState:
		return WrapEntityState(value), nil
	case *corev3.Namespace:
		return WrapNamespace(value), nil
	default:
		return e.etcdWrapFunc(resource, opts...)
	}
}

type WrapList []storev2.Wrapper

func (w WrapList) Unwrap() ([]corev3.Resource, error) {
	result := make([]corev3.Resource, len(w))
	return result, w.UnwrapInto(&result)
}

func (w WrapList) UnwrapInto(dest interface{}) error {
	switch list := dest.(type) {
	case *[]corev3.EntityConfig:
		return w.unwrapIntoEntityConfigList(list)
	case *[]*corev3.EntityConfig:
		return w.unwrapIntoEntityConfigPointerList(list)
	case *[]corev3.EntityState:
		return w.unwrapIntoEntityStateList(list)
	case *[]*corev3.EntityState:
		return w.unwrapIntoEntityStatePointerList(list)
	case *[]corev3.Namespace:
		return w.unwrapIntoNamespaceList(list)
	case *[]*corev3.Namespace:
		return w.unwrapIntoNamespacePointerList(list)
	case *[]corev3.Resource:
		return w.unwrapIntoResourceList(list)
	default:
		return fmt.Errorf("can't unwrap list into %T", dest)
	}
}

func (w WrapList) unwrapIntoEntityConfigList(list *[]corev3.EntityConfig) error {
	if len(*list) != len(w) {
		*list = make([]corev3.EntityConfig, len(w))
	}
	for i, cfg := range w {
		ptr := &((*list)[i])
		if err := cfg.UnwrapInto(ptr); err != nil {
			return err
		}
	}
	return nil
}

func (w WrapList) unwrapIntoEntityConfigPointerList(list *[]*corev3.EntityConfig) error {
	if len(*list) != len(w) {
		*list = make([]*corev3.EntityConfig, len(w))
	}
	for i, cfg := range w {
		ptr := (*list)[i]
		if ptr == nil {
			ptr = new(corev3.EntityConfig)
			(*list)[i] = ptr
		}
		if err := cfg.UnwrapInto(ptr); err != nil {
			return err
		}
	}
	return nil
}

func (w WrapList) unwrapIntoEntityStateList(list *[]corev3.EntityState) error {
	if len(*list) != len(w) {
		*list = make([]corev3.EntityState, len(w))
	}
	for i, state := range w {
		ptr := &((*list)[i])
		if err := state.UnwrapInto(ptr); err != nil {
			return err
		}
	}
	return nil
}

func (w WrapList) unwrapIntoEntityStatePointerList(list *[]*corev3.EntityState) error {
	if len(*list) != len(w) {
		*list = make([]*corev3.EntityState, len(w))
	}
	for i, state := range w {
		ptr := (*list)[i]
		if ptr == nil {
			ptr = new(corev3.EntityState)
			(*list)[i] = ptr
		}
		if err := state.UnwrapInto(ptr); err != nil {
			return err
		}
	}
	return nil
}

func (w WrapList) unwrapIntoNamespaceList(list *[]corev3.Namespace) error {
	if len(*list) != len(w) {
		*list = make([]corev3.Namespace, len(w))
	}
	for i, namespace := range w {
		ptr := &((*list)[i])
		if err := namespace.UnwrapInto(ptr); err != nil {
			return err
		}
	}
	return nil
}

func (w WrapList) unwrapIntoNamespacePointerList(list *[]*corev3.Namespace) error {
	if len(*list) != len(w) {
		*list = make([]*corev3.Namespace, len(w))
	}
	for i, namespace := range w {
		ptr := (*list)[i]
		if ptr == nil {
			ptr = new(corev3.Namespace)
			(*list)[i] = ptr
		}
		if err := namespace.UnwrapInto(ptr); err != nil {
			return err
		}
	}
	return nil
}

// TODO: make this work generically
func (w WrapList) unwrapIntoResourceList(list *[]corev3.Resource) error {
	if len(*list) != len(w) {
		*list = make([]corev3.Resource, len(w))
	}
	for i, resource := range w {
		ptr := (*list)[i]
		if ptr == nil {
			ptr = new(corev3.EntityState)
			(*list)[i] = ptr
		}
		if err := resource.UnwrapInto(ptr); err != nil {
			return err
		}
	}
	return nil
}

func (w WrapList) Len() int {
	return len(w)
}

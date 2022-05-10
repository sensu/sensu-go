package postgres

import (
	"sync"

	corev3 "github.com/sensu/sensu-go/api/core/v3"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/backend/store/v2/wrap"
)

type EnterpriseResourceWrapper struct {
	mu           sync.RWMutex
	etcdWrapFunc func(corev3.Resource, ...wrap.Option) (storev2.Wrapper, error)
	pgEnabled    bool
}

func NewEnterpriseResourceWrapper(wrapFunc func(corev3.Resource, ...wrap.Option) (storev2.Wrapper, error)) *EnterpriseResourceWrapper {
	return &EnterpriseResourceWrapper{
		etcdWrapFunc: wrapFunc,
	}
}

func (e *EnterpriseResourceWrapper) WrapResource(resource corev3.Resource, opts ...wrap.Option) (storev2.Wrapper, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.pgEnabled {
		return e.wrapWithPostgres(resource, opts...)
	}
	return e.etcdWrapFunc(resource, opts...)
}

func (e *EnterpriseResourceWrapper) EnablePostgres() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.pgEnabled = true
}

func (e *EnterpriseResourceWrapper) DisablePostgres() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.pgEnabled = false
}

func (e *EnterpriseResourceWrapper) wrapWithPostgres(resource corev3.Resource, opts ...wrap.Option) (storev2.Wrapper, error) {
	switch value := resource.(type) {
	case *corev3.EntityState:
		return WrapEntityState(value), nil
	default:
		return e.etcdWrapFunc(resource, opts...)
	}
}

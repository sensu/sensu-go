package etcd

import (
	"context"
	"errors"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/traces"
)

const (
	checksPathPrefix = "checks"
)

var (
	checkKeyBuilder = store.NewKeyBuilder(checksPathPrefix)
)

func getCheckConfigPath(check *corev2.CheckConfig) string {
	return checkKeyBuilder.WithResource(check).Build(check.Name)
}

// GetCheckConfigsPath gets the path of the check config store.
func GetCheckConfigsPath(ctx context.Context, name string) string {
	return checkKeyBuilder.WithContext(ctx).Build(name)
}

func schedulerFor(c *corev2.CheckConfig) string {
	if c.Scheduler == "" {
		if c.RoundRobin {
			return corev2.EtcdScheduler
		} else {
			return corev2.MemoryScheduler
		}
	}
	return c.Scheduler
}

// DeleteCheckConfigByName deletes a CheckConfig by name.
func (s *Store) DeleteCheckConfigByName(ctx context.Context, name string) error {
	if name == "" {
		return &store.ErrNotValid{Err: errors.New("must specify name")}
	}

	traceCtx, span := traces.NestedSpan(ctx, "etcd-DeleteCheckConfigByName")
	defer span.End()

	err := Delete(traceCtx, s.client, GetCheckConfigsPath(ctx, name))
	if _, ok := err.(*store.ErrNotFound); ok {
		err = nil
	}
	return err
}

// GetCheckConfigs returns check configurations for an (optional) namespace.
func (s *Store) GetCheckConfigs(ctx context.Context, pred *store.SelectionPredicate) ([]*corev2.CheckConfig, error) {
	checks := []*corev2.CheckConfig{}
	traceCtx, span := traces.NestedSpan(ctx, "etcd-GetCheckConfigs")
	defer span.End()

	err := List(traceCtx, s.client, GetCheckConfigsPath, &checks, pred)
	if err != nil {
		return nil, err
	}
	for _, check := range checks {
		check.Scheduler = schedulerFor(check)
	}
	return checks, err
}

// GetCheckConfigByName gets a CheckConfig by name.
func (s *Store) GetCheckConfigByName(ctx context.Context, name string) (*corev2.CheckConfig, error) {
	if name == "" {
		return nil, &store.ErrNotValid{Err: errors.New("must specify name")}
	}

	traceCtx, span := traces.NestedSpan(ctx, "etcd-GetCheckConfigByName")
	defer span.End()

	var check corev2.CheckConfig
	if err := Get(traceCtx, s.client, GetCheckConfigsPath(traceCtx, name), &check); err != nil {
		if _, ok := err.(*store.ErrNotFound); ok {
			err = nil
		}
		return nil, err
	}
	check.Scheduler = schedulerFor(&check)
	if check.Labels == nil {
		check.Labels = make(map[string]string)
	}
	if check.Annotations == nil {
		check.Annotations = make(map[string]string)
	}

	return &check, nil
}

// UpdateCheckConfig updates a CheckConfig.
func (s *Store) UpdateCheckConfig(ctx context.Context, check *corev2.CheckConfig) error {
	if err := check.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	return CreateOrUpdate(ctx, s.client, getCheckConfigPath(check), check.Namespace, check)
}

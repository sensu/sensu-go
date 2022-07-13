package api

import (
	"context"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

// HookConfigClient is an API client for check hooks.
type HookConfigClient struct {
	client GenericClient
	auth   authorization.Authorizer
}

// NewHookConfigClient creates a new HookConfigClient, given a store and authorizer.
func NewHookConfigClient(store storev2.Interface, auth authorization.Authorizer) *HookConfigClient {
	return &HookConfigClient{
		client: GenericClient{
			Kind:       &corev2.HookConfig{},
			Store:      store,
			Auth:       auth,
			APIGroup:   "core",
			APIVersion: "v2",
		},
		auth: auth,
	}
}

// ListHookConfigs fetches a list of hook resources, if authorized.
func (a *HookConfigClient) ListHookConfigs(ctx context.Context) ([]*corev2.HookConfig, error) {
	pred := &store.SelectionPredicate{
		Continue: corev2.PageContinueFromContext(ctx),
		Limit:    int64(corev2.PageSizeFromContext(ctx)),
	}
	slice := []*corev2.HookConfig{}
	if err := a.client.List(ctx, &slice, pred); err != nil {
		return nil, err
	}
	return slice, nil
}

// FetchHookConfig fetches a hook resource from the backend, if authorized.
func (a *HookConfigClient) FetchHookConfig(ctx context.Context, name string) (*corev2.HookConfig, error) {
	var hook corev2.HookConfig
	if err := a.client.Get(ctx, name, &hook); err != nil {
		return nil, err
	}
	return &hook, nil
}

// CreateHookConfig creates a hook resource, if authorized.
func (a *HookConfigClient) CreateHookConfig(ctx context.Context, hook *corev2.HookConfig) error {
	if err := a.client.Create(ctx, hook); err != nil {
		return err
	}
	return nil
}

// UpdateHookConfig updates a hook resource, if authorized.
func (a *HookConfigClient) UpdateHookConfig(ctx context.Context, hook *corev2.HookConfig) error {
	if err := a.client.Update(ctx, hook); err != nil {
		return err
	}
	return nil
}

// DeleteHookConfig deletes a hook resource, if authorized.
func (a *HookConfigClient) DeleteHookConfig(ctx context.Context, name string) error {
	if err := a.client.Delete(ctx, name); err != nil {
		return err
	}
	return nil
}

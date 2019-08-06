package api

import (
	"context"
	"fmt"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
)

// HookConfigClient is an API client for check hooks.
type HookConfigClient struct {
	client genericClient
	auth   authorization.Authorizer
}

// NewHookConfigClient creates a new HookConfigClient, given a store and authorizer.
func NewHookConfigClient(store store.ResourceStore, auth authorization.Authorizer) *HookConfigClient {
	return &HookConfigClient{
		client: genericClient{
			Kind:       &corev2.HookConfig{},
			Store:      store,
			Auth:       auth,
			Resource:   "hooks",
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
		return nil, fmt.Errorf("couldn't list hooks: %s", err)
	}
	return slice, nil
}

// FetchHookConfig fetches a hook resource from the backend, if authorized.
func (a *HookConfigClient) FetchHookConfig(ctx context.Context, name string) (*corev2.HookConfig, error) {
	var hook corev2.HookConfig
	if err := a.client.Get(ctx, name, &hook); err != nil {
		return nil, fmt.Errorf("couldn't get hook: %s", err)
	}
	return &hook, nil
}

// CreateHookConfig creates a hook resource, if authorized.
func (a *HookConfigClient) CreateHookConfig(ctx context.Context, hook *corev2.HookConfig) error {
	if err := a.client.Create(ctx, hook); err != nil {
		return fmt.Errorf("couldn't create hook: %s", err)
	}
	return nil
}

// UpdateHookConfig updates a hook resource, if authorized.
func (a *HookConfigClient) UpdateHookConfig(ctx context.Context, hook *corev2.HookConfig) error {
	if err := a.client.Update(ctx, hook); err != nil {
		return fmt.Errorf("couldn't update hook: %s", err)
	}
	return nil
}

package api

import (
	"context"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
)

// CheckController represents the controller needs of the ChecksRouter.
type CheckController interface {
	AddCheckHook(context.Context, string, corev2.HookList) error
	RemoveCheckHook(context.Context, string, string, string) error
	QueueAdhocRequest(context.Context, string, *corev2.AdhocRequest) error
}

// CheckClient is an API client for check configuration.
type CheckClient struct {
	store      store.CheckConfigStore
	controller CheckController
	auth       authorization.Authorizer
}

// NewCheckClient creates a new CheckClient, given a store, a controller, and
// an authorizer.
func NewCheckClient(store store.Store, controller CheckController, auth authorization.Authorizer) *CheckClient {
	return &CheckClient{store: store, controller: controller, auth: auth}
}

// CreateCheck creates a new check, if authorized.
func (c *CheckClient) CreateCheck(ctx context.Context, check *corev2.CheckConfig) error {
	attrs := checkCreateAttributes(ctx, check.Name)
	if err := authorize(ctx, c.auth, attrs); err != nil {
		return err
	}
	setCreatedBy(ctx, check)
	return c.store.UpdateCheckConfig(ctx, check)
}

// UpdateCheck updates a check, if authorized.
func (c *CheckClient) UpdateCheck(ctx context.Context, check *corev2.CheckConfig) error {
	attrs := checkUpdateAttributes(ctx, check.Name)
	if err := authorize(ctx, c.auth, attrs); err != nil {
		return err
	}
	setCreatedBy(ctx, check)
	return c.store.UpdateCheckConfig(ctx, check)
}

// DeleteCheck deletes a check, if authorized.
func (c *CheckClient) DeleteCheck(ctx context.Context, name string) error {
	attrs := checkDeleteAttributes(ctx, name)
	if err := authorize(ctx, c.auth, attrs); err != nil {
		return err
	}
	return c.store.DeleteCheckConfigByName(ctx, name)
}

// ExecuteCheck queues an ahoc check request, if authorized.
func (c *CheckClient) ExecuteCheck(ctx context.Context, name string, req *corev2.AdhocRequest) error {
	attrs := checkCreateAttributes(ctx, name)
	if err := authorize(ctx, c.auth, attrs); err != nil {
		return err
	}
	return c.controller.QueueAdhocRequest(ctx, name, req)
}

// FetchCheck retrieves a check, if authorized.
func (c *CheckClient) FetchCheck(ctx context.Context, name string) (*corev2.CheckConfig, error) {
	attrs := checkFetchAttributes(ctx, name)
	if err := authorize(ctx, c.auth, attrs); err != nil {
		return nil, err
	}
	return c.store.GetCheckConfigByName(ctx, name)
}

// ListChecks lists all checks in a namespace, if authorized.
func (c *CheckClient) ListChecks(ctx context.Context) ([]*corev2.CheckConfig, error) {
	attrs := checkListAttributes(ctx)
	if err := authorize(ctx, c.auth, attrs); err != nil {
		return nil, err
	}
	pred := &store.SelectionPredicate{
		Continue: corev2.PageContinueFromContext(ctx),
		Limit:    int64(corev2.PageSizeFromContext(ctx)),
	}
	return c.store.GetCheckConfigs(ctx, pred)
}

func checkListAttributes(ctx context.Context) *authorization.Attributes {
	return &authorization.Attributes{
		APIGroup:   "core",
		APIVersion: "v2",
		Namespace:  corev2.ContextNamespace(ctx),
		Resource:   "checks",
		Verb:       "list",
	}
}

func checkFetchAttributes(ctx context.Context, name string) *authorization.Attributes {
	return &authorization.Attributes{
		APIGroup:     "core",
		APIVersion:   "v2",
		Namespace:    corev2.ContextNamespace(ctx),
		Resource:     "checks",
		Verb:         "get",
		ResourceName: name,
	}
}

func checkCreateAttributes(ctx context.Context, name string) *authorization.Attributes {
	return &authorization.Attributes{
		APIGroup:     "core",
		APIVersion:   "v2",
		Namespace:    corev2.ContextNamespace(ctx),
		Resource:     "checks",
		Verb:         "create",
		ResourceName: name,
	}
}

func checkUpdateAttributes(ctx context.Context, name string) *authorization.Attributes {
	return &authorization.Attributes{
		APIGroup:     "core",
		APIVersion:   "v2",
		Namespace:    corev2.ContextNamespace(ctx),
		Resource:     "checks",
		Verb:         "update",
		ResourceName: name,
	}
}

func checkDeleteAttributes(ctx context.Context, name string) *authorization.Attributes {
	return &authorization.Attributes{
		APIGroup:     "core",
		APIVersion:   "v2",
		Namespace:    corev2.ContextNamespace(ctx),
		Resource:     "checks",
		Verb:         "delete",
		ResourceName: name,
	}
}

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

type CheckClient struct {
	store      store.CheckConfigStore
	controller CheckController
	auth       authorization.Authorizer
}

func NewCheckClient(store store.Store, controller CheckController, auth authorization.Authorizer) *CheckClient {
	return &CheckClient{store: store, controller: controller, auth: auth}
}

func (c *CheckClient) CreateCheck(ctx context.Context, check *corev2.CheckConfig) error {
	attrs := checkCreateAttributes(ctx, check.Name)
	if err := authorize(ctx, c.auth, attrs); err != nil {
		return err
	}
	return c.store.UpdateCheckConfig(ctx, check)
}

func (c *CheckClient) UpdateCheck(ctx context.Context, check *corev2.CheckConfig) error {
	attrs := checkUpdateAttributes(ctx, check.Name)
	if err := authorize(ctx, c.auth, attrs); err != nil {
		return err
	}
	return c.store.UpdateCheckConfig(ctx, check)
}

func (c *CheckClient) DeleteCheck(ctx context.Context, name string) error {
	attrs := checkDeleteAttributes(ctx, name)
	if err := authorize(ctx, c.auth, attrs); err != nil {
		return err
	}
	return c.store.DeleteCheckConfigByName(ctx, name)
}

func (c *CheckClient) ExecuteCheck(ctx context.Context, name string, req *corev2.AdhocRequest) error {
	attrs := checkCreateAttributes(ctx, name)
	if err := authorize(ctx, c.auth, attrs); err != nil {
		return err
	}
	return c.controller.QueueAdhocRequest(ctx, name, req)
}

func (c *CheckClient) FetchCheck(ctx context.Context, name string) (*corev2.CheckConfig, error) {
	attrs := checkFetchAttributes(ctx, name)
	if err := authorize(ctx, c.auth, attrs); err != nil {
		return nil, err
	}
	return c.store.GetCheckConfigByName(ctx, name)
}

func (c *CheckClient) ListChecks(ctx context.Context) ([]*corev2.CheckConfig, error) {
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

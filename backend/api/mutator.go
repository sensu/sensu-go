package api

import (
	"context"
	"fmt"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
)

type MutatorClient struct {
	client genericClient
	auth   authorization.Authorizer
}

func NewMutatorClient(store store.ResourceStore, auth authorization.Authorizer) *MutatorClient {
	return &MutatorClient{
		client: genericClient{
			Kind:       &corev2.Mutator{},
			Store:      store,
			Auth:       auth,
			Resource:   "mutators",
			APIGroup:   "core",
			APIVersion: "v2",
		},
		auth: auth,
	}
}

// ListMutators fetches a list of mutator resources
func (a *MutatorClient) ListMutators(ctx context.Context) ([]*corev2.Mutator, error) {
	pred := &store.SelectionPredicate{
		Continue: corev2.PageContinueFromContext(ctx),
		Limit:    int64(corev2.PageSizeFromContext(ctx)),
	}
	slice := []*corev2.Mutator{}
	if err := a.client.List(ctx, &slice, pred); err != nil {
		return nil, fmt.Errorf("couldn't list mutators: %s", err)
	}
	return slice, nil
}

// FetchMutator fetches a mutator resource from the backend
func (a *MutatorClient) FetchMutator(ctx context.Context, name string) (*corev2.Mutator, error) {
	var mutator corev2.Mutator
	if err := a.client.Get(ctx, name, &mutator); err != nil {
		return nil, fmt.Errorf("couldn't get mutator: %s", err)
	}
	return &mutator, nil
}

// CreateMutator creates a mutator resource
func (a *MutatorClient) CreateMutator(ctx context.Context, mutator *corev2.Mutator) error {
	if err := a.client.Create(ctx, mutator); err != nil {
		return fmt.Errorf("couldn't create mutator: %s", err)
	}
	return nil
}

// UpdateMutator updates a mutator resource
func (a *MutatorClient) UpdateMutator(ctx context.Context, mutator *corev2.Mutator) error {
	if err := a.client.Update(ctx, mutator); err != nil {
		return fmt.Errorf("couldn't update mutator: %s", err)
	}
	return nil
}

package api

import (
	"context"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
)

// MutatorClient is an API client for mutators.
type MutatorClient struct {
	client GenericClient
	auth   authorization.Authorizer
}

// NewMutatorClient creates a new MutatorClient, given a store and an authorizer.
func NewMutatorClient(store store.ResourceStore, auth authorization.Authorizer) *MutatorClient {
	return &MutatorClient{
		client: GenericClient{
			Kind:       &corev2.Mutator{},
			Store:      store,
			Auth:       auth,
			APIGroup:   "core",
			APIVersion: "v2",
		},
		auth: auth,
	}
}

// ListMutators fetches a list of mutator resources, if authorized.
func (a *MutatorClient) ListMutators(ctx context.Context) ([]*corev2.Mutator, error) {
	pred := &store.SelectionPredicate{
		Continue: corev2.PageContinueFromContext(ctx),
		Limit:    int64(corev2.PageSizeFromContext(ctx)),
	}
	slice := []*corev2.Mutator{}
	if err := a.client.List(ctx, &slice, pred); err != nil {
		return nil, err
	}
	return slice, nil
}

// FetchMutator fetches a mutator resource from the backend, if authorized.
func (a *MutatorClient) FetchMutator(ctx context.Context, name string) (*corev2.Mutator, error) {
	var mutator corev2.Mutator
	if err := a.client.Get(ctx, name, &mutator); err != nil {
		return nil, err
	}
	return &mutator, nil
}

// CreateMutator creates a mutator resource, if authorized.
func (a *MutatorClient) CreateMutator(ctx context.Context, mutator *corev2.Mutator) error {
	if err := a.client.Create(ctx, mutator); err != nil {
		return err
	}
	return nil
}

// UpdateMutator updates a mutator resource, if authorized.
func (a *MutatorClient) UpdateMutator(ctx context.Context, mutator *corev2.Mutator) error {
	if err := a.client.Update(ctx, mutator); err != nil {
		return err
	}
	return nil
}

// DeleteMutator deletes a mutator resource, if authorized.
func (a *MutatorClient) DeleteMutator(ctx context.Context, name string) error {
	if err := a.client.Delete(ctx, name); err != nil {
		return err
	}
	return nil
}

package api

import (
	"context"
	"fmt"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

// SilencedClient is an API client for silencing checks.
type SilencedClient struct {
	store storev2.SilencesStore
	auth  authorization.Authorizer
}

// NewSilencedClient creates a new SilencedClient, given a store and authorizer.
func NewSilencedClient(store store.SilenceStore, auth authorization.Authorizer) *SilencedClient {
	return &SilencedClient{
		store: store,
		auth:  auth,
	}
}

// UpdateSilenced updates a silenced entry, if authorized.
func (s *SilencedClient) UpdateSilenced(ctx context.Context, silenced *corev2.Silenced) error {
	silenced.Prepare(ctx)
	if err := silenced.Validate(); err != nil {
		return fmt.Errorf("couldn't update silenced entry: %s", err)
	}
	attrs := silencedUpdateAttrs(ctx, silenced.Name)
	if err := authorize(ctx, s.auth, attrs); err != nil {
		return err
	}
	setCreatedBy(ctx, silenced)
	if err := s.store.UpdateSilence(ctx, silenced); err != nil {
		return fmt.Errorf("couldn't update silenced entry: %s", err)
	}
	return nil
}

// GetSilencedByName gets a silenced entry by name, if authorized.
func (s *SilencedClient) GetSilencedByName(ctx context.Context, name string) (*corev2.Silenced, error) {
	attrs := silencedFetchAttrs(ctx, name)
	if err := authorize(ctx, s.auth, attrs); err != nil {
		return nil, err
	}
	silenced, err := s.store.GetSilenceByName(ctx, corev2.ContextNamespace(ctx), name)
	if err != nil {
		return nil, fmt.Errorf("couldn't get silenced entry: %s", err)
	}
	return silenced, nil
}

// DeleteSilencedByName deletes a silenced entry by name, if authorized.
func (s *SilencedClient) DeleteSilencedByName(ctx context.Context, name string) error {
	attrs := silencedDeleteAttrs(ctx, name)
	if err := authorize(ctx, s.auth, attrs); err != nil {
		return err
	}
	if err := s.store.DeleteSilences(ctx, corev2.ContextNamespace(ctx), []string{name}); err != nil {
		return fmt.Errorf("couldn't delete silenced entry: %s", err)
	}
	return nil
}

// ListSilenced lists all silenced entries within a namespace, if authorized.
func (s *SilencedClient) ListSilenced(ctx context.Context) ([]*corev2.Silenced, error) {
	attrs := silencedListAttrs(ctx)
	if err := authorize(ctx, s.auth, attrs); err != nil {
		return nil, err
	}
	silenceds, err := s.store.GetSilences(ctx, corev2.ContextNamespace(ctx))
	if err != nil {
		return nil, fmt.Errorf("couldn't list silenced entries: %s", err)
	}
	return silenceds, nil
}

// GetSilencedByCheckName gets all of the silenced entries applied to a check,
// if authorized.
func (s *SilencedClient) GetSilencedByCheckName(ctx context.Context, check string) ([]*corev2.Silenced, error) {
	// access to the check implies access to its silenced entries
	attrs := checkFetchAttributes(ctx, check)
	if err := authorize(ctx, s.auth, attrs); err != nil {
		return nil, err
	}
	silenceds, err := s.store.GetSilencesByCheck(ctx, corev2.ContextNamespace(ctx), check)
	if err != nil {
		return nil, fmt.Errorf("couldn't list silenced entries: %s", err)
	}
	return silenceds, nil
}

// GetSilencedBySubscription gets all of the silenced entries applied to a
// subscription, if authorized.
func (s *SilencedClient) GetSilencedBySubscription(ctx context.Context, subs ...string) ([]*corev2.Silenced, error) {
	attrs := silencedListAttrs(ctx)
	if err := authorize(ctx, s.auth, attrs); err != nil {
		return nil, err
	}
	silenceds, err := s.store.GetSilencesBySubscription(ctx, corev2.ContextNamespace(ctx), subs)
	if err != nil {
		return nil, fmt.Errorf("couldn't list silenced entries: %s", err)
	}
	return silenceds, nil
}

func silencedUpdateAttrs(ctx context.Context, name string) *authorization.Attributes {
	return &authorization.Attributes{
		APIGroup:     "core",
		APIVersion:   "v2",
		Namespace:    corev2.ContextNamespace(ctx),
		Resource:     "silenced",
		Verb:         "update",
		ResourceName: name,
	}
}

func silencedFetchAttrs(ctx context.Context, name string) *authorization.Attributes {
	return &authorization.Attributes{
		APIGroup:     "core",
		APIVersion:   "v2",
		Namespace:    corev2.ContextNamespace(ctx),
		Resource:     "silenced",
		Verb:         "get",
		ResourceName: name,
	}
}

func silencedDeleteAttrs(ctx context.Context, name string) *authorization.Attributes {
	return &authorization.Attributes{
		APIGroup:     "core",
		APIVersion:   "v2",
		Namespace:    corev2.ContextNamespace(ctx),
		Resource:     "silenced",
		Verb:         "delete",
		ResourceName: name,
	}
}

func silencedListAttrs(ctx context.Context) *authorization.Attributes {
	return &authorization.Attributes{
		APIGroup:   "core",
		APIVersion: "v2",
		Namespace:  corev2.ContextNamespace(ctx),
		Resource:   "silenced",
		Verb:       "list",
	}
}

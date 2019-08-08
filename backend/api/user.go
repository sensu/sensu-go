package api

import (
	"context"
	"fmt"

	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// UserClient is an API client for users.
type UserClient struct {
	client GenericClient
	auth   authorization.Authorizer
}

// NewUserClient creates a new UserClient, given a store and an authorizer.
func NewUserClient(store store.ResourceStore, auth authorization.Authorizer) *UserClient {
	return &UserClient{
		client: GenericClient{
			Kind:       &corev2.User{},
			Store:      store,
			Auth:       auth,
			APIGroup:   "core",
			APIVersion: "v2",
		},
		auth: auth,
	}
}

// ListUsers fetches a list of user resources, if authorized.
func (a *UserClient) ListUsers(ctx context.Context) ([]*corev2.User, error) {
	pred := &store.SelectionPredicate{
		Continue: corev2.PageContinueFromContext(ctx),
		Limit:    int64(corev2.PageSizeFromContext(ctx)),
	}
	slice := []*corev2.User{}
	if err := a.client.List(ctx, &slice, pred); err != nil {
		return nil, fmt.Errorf("couldn't list users: %s", err)
	}
	return slice, nil
}

// FetchUser fetches an user resource from the backend, if authorized.
func (a *UserClient) FetchUser(ctx context.Context, name string) (*corev2.User, error) {
	var user corev2.User
	if err := a.client.Get(ctx, name, &user); err != nil {
		return nil, fmt.Errorf("couldn't get user: %s", err)
	}
	return &user, nil
}

// CreateUser creates an user resource, if authorized.
func (a *UserClient) CreateUser(ctx context.Context, user *corev2.User) error {
	if err := a.client.Create(ctx, user); err != nil {
		return fmt.Errorf("couldn't create user: %s", err)
	}
	return nil
}

// UpdateUser updates an user resource, if authorized.
func (a *UserClient) UpdateUser(ctx context.Context, user *corev2.User) error {
	if err := a.client.Update(ctx, user); err != nil {
		return fmt.Errorf("couldn't update user: %s", err)
	}
	return nil
}

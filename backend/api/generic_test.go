package api

import (
	"context"
	"errors"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/mock"
)

func defaultTestClient(store store.ResourceStore, auth authorization.Authorizer) *GenericClient {
	return &GenericClient{
		Kind:       defaultResource(),
		Store:      store,
		Auth:       auth,
		APIGroup:   "core",
		APIVersion: "v2",
	}
}

type mockAuth struct {
	attrs map[authorization.AttributesKey]bool
}

func defaultResource() corev2.Resource {
	return corev2.FixtureAsset("default")
}

func (m *mockAuth) Authorize(ctx context.Context, attrs *authorization.Attributes) (bool, error) {
	if attrs == nil {
		return false, nil
	}
	if corev2.ContextNamespace(ctx) != attrs.Namespace {
		return false, errors.New("namespace does not match")
	}
	b, ok := m.attrs[attrs.Key()]
	if !ok {
		return false, errors.New("no information")
	}
	return b, nil
}

func badAuth() authorization.Authorizer {
	return &mockAuth{}
}

func contextWithUser(ctx context.Context, username string, groups []string) context.Context {
	return context.WithValue(ctx, corev2.ClaimsKey, corev2.FixtureClaims(username, groups))
}

func defaultResourceStore() store.ResourceStore {
	store := &mockstore.MockStore{}
	store.On("GetResource", mock.Anything, "default", mock.Anything).Run(func(args mock.Arguments) {
		arg := args.Get(2).(*corev2.Asset)
		*arg = *corev2.FixtureAsset("default")
	}).Return(nil)
	store.On("ListResources", mock.Anything, (&corev2.Asset{}).StorePrefix(), mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		arg := args.Get(2).(*[]corev2.Resource)
		*arg = []corev2.Resource{defaultResource()}
	}).Return(nil)
	store.On("CreateResource", mock.Anything, mock.Anything).Return(nil)
	store.On("CreateOrUpdateResource", mock.Anything, mock.Anything).Return(nil)
	store.On("DeleteResource", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	return store
}

func TestGenericClient(t *testing.T) {
	tests := []struct {
		Name      string
		Client    *GenericClient
		CreateVal corev2.Resource
		CreateErr bool
		UpdateVal corev2.Resource
		UpdateErr bool
		GetName   string
		GetVal    corev2.Resource
		GetErr    bool
		ListVal   []corev2.Resource
		ListPred  *store.SelectionPredicate
		ListErr   bool
		DelName   string
		DelErr    bool
		Ctx       context.Context
	}{
		{
			Name:      "authorizer rejects all",
			Client:    defaultTestClient(defaultResourceStore(), badAuth()),
			CreateVal: defaultResource(),
			CreateErr: true,
			UpdateVal: defaultResource(),
			UpdateErr: true,
			GetName:   "toget",
			GetVal:    defaultResource(),
			GetErr:    true,
			DelName:   "todelete",
			DelErr:    true,
			ListErr:   true,
			Ctx:       contextWithUser(defaultContext(), "tom", nil),
		},
		{
			Name: "readonly access",
			Client: defaultTestClient(defaultResourceStore(), &mockAuth{
				attrs: map[authorization.AttributesKey]bool{
					authorization.AttributesKey{
						APIGroup:     "core",
						APIVersion:   "v2",
						Namespace:    "default",
						Resource:     "assets",
						ResourceName: "default",
						UserName:     "tom",
						Verb:         "get",
					}: true,
					authorization.AttributesKey{
						APIGroup:   "core",
						APIVersion: "v2",
						Namespace:  "default",
						Resource:   "assets",
						UserName:   "tom",
						Verb:       "list",
					}: true,
				},
			}),
			CreateVal: defaultResource(),
			CreateErr: true,
			UpdateVal: defaultResource(),
			UpdateErr: true,
			GetName:   "default",
			GetVal:    defaultResource(),
			GetErr:    false,
			DelName:   "default",
			DelErr:    true,
			ListErr:   false,
			Ctx:       contextWithUser(defaultContext(), "tom", nil),
		},
		{
			Name: "all access",
			Client: defaultTestClient(defaultResourceStore(), &mockAuth{
				attrs: map[authorization.AttributesKey]bool{
					authorization.AttributesKey{
						APIGroup:     "core",
						APIVersion:   "v2",
						Namespace:    "default",
						Resource:     "assets",
						ResourceName: "default",
						UserName:     "tom",
						Verb:         "get",
					}: true,
					authorization.AttributesKey{
						APIGroup:     "core",
						APIVersion:   "v2",
						Namespace:    "default",
						Resource:     "assets",
						ResourceName: "default",
						UserName:     "tom",
						Verb:         "create",
					}: true,
					authorization.AttributesKey{
						APIGroup:     "core",
						APIVersion:   "v2",
						Namespace:    "default",
						Resource:     "assets",
						ResourceName: "default",
						UserName:     "tom",
						Verb:         "delete",
					}: true,
					authorization.AttributesKey{
						APIGroup:     "core",
						APIVersion:   "v2",
						Namespace:    "default",
						Resource:     "assets",
						ResourceName: "default",
						UserName:     "tom",
						Verb:         "update",
					}: true,
					authorization.AttributesKey{
						APIGroup:   "core",
						APIVersion: "v2",
						Namespace:  "default",
						Resource:   "assets",
						UserName:   "tom",
						Verb:       "list",
					}: true,
				},
			}),
			CreateVal: defaultResource(),
			CreateErr: false,
			UpdateVal: defaultResource(),
			UpdateErr: false,
			GetName:   "default",
			GetVal:    defaultResource(),
			GetErr:    false,
			DelName:   "default",
			DelErr:    false,
			ListErr:   false,
			Ctx:       contextWithUser(defaultContext(), "tom", nil),
		},
	}

	for _, test := range tests {
		t.Run(test.Name+"_create", func(t *testing.T) {
			err := test.Client.Create(test.Ctx, test.CreateVal)
			if err != nil && !test.CreateErr {
				t.Fatal(err)
			}
			if err == nil && test.CreateErr {
				t.Fatal("expected non-nil error")
			}
		})
		t.Run(test.Name+"_update", func(t *testing.T) {
			err := test.Client.Update(test.Ctx, test.UpdateVal)
			if err != nil && !test.UpdateErr {
				t.Fatal(err)
			}
			if err == nil && test.UpdateErr {
				t.Fatal(err)
			}
		})
		t.Run(test.Name+"_get", func(t *testing.T) {
			err := test.Client.Get(test.Ctx, test.GetName, test.GetVal)
			if err != nil && !test.GetErr {
				t.Fatal(err)
			}
			if err == nil && test.GetErr {
				t.Fatal(err)
			}
			if err == nil && test.GetVal.Validate() != nil {
				t.Fatal(test.GetVal.Validate())
			}
		})
		t.Run(test.Name+"_del", func(t *testing.T) {
			err := test.Client.Delete(test.Ctx, test.DelName)
			if err != nil && !test.DelErr {
				t.Fatal(err)
			}
			if err == nil && test.DelErr {
				t.Fatal(err)
			}
		})
		t.Run(test.Name+"_list", func(t *testing.T) {
			err := test.Client.List(test.Ctx, &test.ListVal, test.ListPred)
			if err != nil && !test.ListErr {
				t.Fatal(err)
			}
			if err == nil && test.ListErr {
				t.Fatal(err)
			}
			for _, val := range test.ListVal {
				if err == nil && val.Validate() != nil {
					t.Fatal(val.Validate())
				}
			}
		})
	}
}
